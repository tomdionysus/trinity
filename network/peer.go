package network

import (
	"crypto/tls"
	"errors"
	"github.com/tomdionysus/consistenthash"
	"github.com/tomdionysus/trinity/packets"
	"github.com/tomdionysus/trinity/util"
	// "bytes"
	"encoding/gob"
	"strings"
	"time"
)

const (
	PeerStateDisconnected = iota
	PeerStateConnecting   = iota
	PeerStateHandshake    = iota
	PeerStateConnected    = iota
	PeerStateSyncing      = iota
	PeerStateDefib        = iota
)

var PeerStateString map[uint]string = map[uint]string{
	PeerStateDisconnected: "PeerStateDisconnected",
	PeerStateConnecting:   "PeerStateConnecting",
	PeerStateHandshake:    "PeerStateHandshake",
	PeerStateConnected:    "PeerStateConnected",
	PeerStateSyncing:      "PeerStateSyncing",
	PeerStateDefib:        "PeerStateDefib",
}

type Peer struct {
	Logger *util.Logger
	Server *TLSServer

	Incoming bool

	Address string
	State   uint

	Connection      *tls.Conn
	HeartbeatTicker *time.Ticker

	Writer *gob.Encoder
	Reader *gob.Decoder

	LastHeartbeat time.Time

	ServerNetworkNode *consistenthash.ServerNetworkNode

	Replies map[[16]byte]chan (*packets.Packet)
}

func NewPeer(logger *util.Logger, server *TLSServer, address string) *Peer {
	inst := &Peer{
		Logger:            logger,
		Address:           address,
		State:             PeerStateDisconnected,
		Server:            server,
		LastHeartbeat:     time.Now(),
		ServerNetworkNode: nil,
		Replies:           map[[16]byte]chan (*packets.Packet){},
	}
	return inst
}

func NewConnectingPeer(logger *util.Logger, server *TLSServer, connection *tls.Conn) *Peer {
	inst := NewPeer(logger, server, connection.RemoteAddr().String())
	inst.Connection = connection
	inst.State = PeerStateHandshake
	inst.Incoming = true
	return inst
}

func (pr *Peer) Connect() error {
	pr.Incoming = false
	pr.State = PeerStateConnecting
	conn, err := tls.Dial("tcp", pr.Address, &tls.Config{
		RootCAs:      pr.Server.CAPool.Pool,
		Certificates: []tls.Certificate{*pr.Server.Certificate},
	})
	if err != nil {
		pr.Logger.Error("Peer", "Cannot connect to %s: %s", pr.Address, err.Error())
		pr.Disconnect()
		return err
	}
	pr.Connection = conn
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		pr.Logger.Error("Peer", "Cannot connect to %s: Peer has no certificates", pr.Address)
		pr.Disconnect()
		return errors.New("Peer has no certificates")
	}
	pr.State = PeerStateHandshake
	return nil
}

func (pr *Peer) Disconnect() {
	if pr.State != PeerStateDisconnected {
		pr.State = PeerStateDisconnected
		pr.Server.ServerNode.DeregisterNode(pr.ServerNetworkNode)
		if pr.HeartbeatTicker != nil {
			pr.HeartbeatTicker.Stop()
		}
		if pr.Connection != nil {
			pr.Connection.Close()
		}
		pr.Logger.Info("Peer", "%02X: Disconnected", pr.ServerNetworkNode.ID)
		delete(pr.Server.Connections, pr.ServerNetworkNode.ID)
	}
}

func (pr *Peer) Start() error {
	if pr.State != PeerStateHandshake {
		pr.Logger.Error("Peer", "Cannot Start Client, Handshake not ready")
		return errors.New("Handshake not ready")
	}
	err := pr.Connection.Handshake()
	if err != nil {
		pr.Logger.Error("Peer", "Peer TLS Handshake failed, disconnecting: %s", err.Error())
		pr.Disconnect()
		return errors.New("TLS Handshake Failed")
	}
	state := pr.Connection.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		pr.Logger.Error("Peer", "Peer has no certificates, disconnecting")
		pr.Disconnect()
		return errors.New("Peer sent no certificates")
	}
	sub := state.PeerCertificates[0].Subject.CommonName

	if pr.Incoming {
		pr.Logger.Info("Peer", "Outgoing Connection to %s (%s) [%s]", pr.Connection.RemoteAddr(), sub, Ciphers[pr.Connection.ConnectionState().CipherSuite])
	} else {
		pr.Logger.Info("Peer", "Incoming Connection from %s (%s) [%s]", pr.Connection.RemoteAddr(), sub, Ciphers[pr.Connection.ConnectionState().CipherSuite])
	}

	pr.Reader = gob.NewDecoder(pr.Connection)
	pr.Writer = gob.NewEncoder(pr.Connection)

	go pr.heartbeat()

	pr.SendDistribution()

	go pr.process()

	return nil
}

// Ping the Peer every second.
func (pr *Peer) heartbeat() {
	pr.HeartbeatTicker = time.NewTicker(time.Second)

	for {
		<-pr.HeartbeatTicker.C

		// Check For Defib
		if time.Now().After(pr.LastHeartbeat.Add(5 * time.Second)) {
			pr.Logger.Warn("Peer", "%02X: Peer Defib (no response for >5 seconds)", pr.ServerNetworkNode.ID)
			pr.State = PeerStateDefib
		}

		switch pr.State {
		case PeerStateConnected:
			err := pr.SendPacket(packets.NewPacket(packets.CMD_HEARTBEAT, nil))
			if err != nil {
				pr.Logger.Error("Peer", "%02X: Error Sending Heartbeat, disconnecting", pr.ServerNetworkNode.ID)
				pr.Disconnect()
				return
			}
		case PeerStateDefib:
			if time.Now().After(pr.LastHeartbeat.Add(10 * time.Second)) {
				pr.Logger.Warn("Peer", "%02X: Peer DOA (Defib for 5 seconds, disconnecting)", pr.ServerNetworkNode.ID)
				pr.Disconnect()
				return
			}
		}

	}

}

func (pr *Peer) process() {

	var packet packets.Packet

	for {

		// Read Command
		err := pr.Reader.Decode(&packet)
		if err != nil {
			if err.Error() == "EOF" {
				pr.Logger.Debug("Peer", "%02X: Peer Closed Connection", pr.ServerNetworkNode.ID)
			} else {
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					pr.Logger.Debug("Peer", "%02X: Read After This Node Closed Connection", pr.ServerNetworkNode.ID)
				} else {
					pr.Logger.Error("Peer", "%02X: Error Reading: %s", pr.ServerNetworkNode.ID, err.Error())
				}
			}
			goto end
		}
		switch packet.Command {

		// Packets in Connecting / Handshake

		case packets.CMD_HEARTBEAT:
			pr.LastHeartbeat = time.Now()

		case packets.CMD_DISTRIBUTION:
			if pr.ServerNetworkNode != nil {
				pr.Logger.Warn("Peer", "%02X: CMD_DISTRIBUTION received from registered peer", pr.ServerNetworkNode.ID)
				break
			}
			servernetworknode := packet.Payload.(consistenthash.ServerNetworkNode)
			pr.ServerNetworkNode = &servernetworknode
			pr.Server.Connections[pr.ServerNetworkNode.ID] = pr
			pr.Logger.Debug("Peer", "%02X: CMD_DISTRIBUTION (%s)", pr.ServerNetworkNode.ID, pr.Connection.RemoteAddr())

			if pr.Server.ServerNode.ID == pr.ServerNetworkNode.ID {
				if !pr.Incoming {
					pr.Logger.Warn("Peer", "%02X: The outgoing connection connected back to this node - disconnecting (%s)", pr.ServerNetworkNode.ID, pr.Connection.RemoteAddr())
				} else {
					pr.Logger.Debug("Peer", "%02X: Incoming connection is from this node - disconnecting (%s)", pr.ServerNetworkNode.ID, pr.Connection.RemoteAddr())
				}
				pr.Disconnect()
				break
			}

			if pr.Server.ServerNode.NodeRegistered(pr.ServerNetworkNode.ID) {
				// Peer has previously registered
				pr.Logger.Debug("Peer", "%02X: Node Already Registered", pr.ServerNetworkNode.ID)
			} else {
				// Peer is new
				_, err := pr.Server.ServerNode.RegisterNode(pr.ServerNetworkNode)
				if err != nil {
					pr.Logger.Error("Peer", "%02X: Register Node Distribution Failed: %s", pr.ServerNetworkNode.ID, err.Error())
					break
				}
			}

			pr.State = PeerStateConnected
			pr.LastHeartbeat = time.Now()
			
			pr.Server.NotifyAllPeers()

		// Packets in Connected

		case packets.CMD_KVSTORE:
			pr.Logger.Debug("Peer", "%02X: CMD_KVSTORE", pr.ServerNetworkNode.ID)
			pr.handleKVStorePacket(&packet)

		case packets.CMD_KVSTORE_ACK:
			pr.Logger.Debug("Peer", "%02X: CMD_KVSTORE_ACK", pr.ServerNetworkNode.ID)
			pr.handleReply(&packet)

		case packets.CMD_KVSTORE_NOT_FOUND:
			pr.Logger.Debug("Peer", "%02X: CMD_KVSTORE_NOT_FOUND", pr.ServerNetworkNode.ID)
			pr.handleReply(&packet)

		case packets.CMD_PEERLIST:
			peers := packet.Payload.(packets.PeerListPacket)
			pr.Logger.Debug("Peer", "%02X: CMD_PEERLIST (%d Peers)", pr.ServerNetworkNode.ID, len(peers))

			for id, _ :=range pr.Server.Connections {
					pr.Logger.Debug("Peer", "CMD_PEERLIST Listing Existing Peer %02X",id)
			}

			for id, k := range peers {
				if pr.Server.ServerNode.ID == id {
					pr.Logger.Warn("Peer", "%02X: - Notified us of ourselves (%02X).", pr.ServerNetworkNode.ID, id)
				} else {
					if !pr.Server.IsConnectedTo(id) {
						pr.Logger.Debug("Peer", "%02X: - Notified %02X - Connecting New Peer (%s)", pr.ServerNetworkNode.ID, id, k)
						pr.Server.ConnectTo(k)
					} else {
						pr.Logger.Debug("Peer", "%02X: - Notified %02X - Already Connected to Peer (%s)", pr.ServerNetworkNode.ID, id, k)
					}
				}
			}

		default:
			pr.Logger.Warn("Peer", "%02X: Unknown Packet Command %d", pr.ServerNetworkNode.ID, packet.Command)
		}

		if pr.State == PeerStateDisconnected {
			break
		}
	}
end:

	pr.Disconnect()
}

func (pr *Peer) handleReply(packet *packets.Packet) {
	chn, found := pr.Replies[packet.RequestID]
	if found {
		delete(pr.Replies, packet.RequestID)
		chn <- packet
	} else {
		pr.Logger.Warn("Peer", "%02X: Unsolicited Reply to unknown packet %02X", pr.ServerNetworkNode.ID, packet.RequestID)
	}
}

func (pr *Peer) SendDistribution() error {
	packet := packets.NewPacket(packets.CMD_DISTRIBUTION, pr.Server.ServerNode.ServerNetworkNode)
	pr.SendPacket(packet)
	return nil
}

func (pr *Peer) SendPacket(packet *packets.Packet) error {
	err := pr.Writer.Encode(packet)
	if err != nil {
		pr.Logger.Error("Peer", "Error Writing: %s", err.Error())
	}
	return err
}

func (pr *Peer) SendPacketWaitReply(packet *packets.Packet, timeout time.Duration) (*packets.Packet, error) {
	if pr.State != PeerStateConnected {
		pr.Logger.Error("Peer", "%02X: Cannot send packet ID %02X, not PeerStateConnected", pr.ServerNetworkNode.ID, packet.ID)
		return nil, errors.New("Cannot send, state not PeerStateConnected")
	}

	pr.Replies[packet.ID] = make(chan (*packets.Packet))
	ticker := time.NewTicker(timeout)
	pr.SendPacket(packet)
	var reply *packets.Packet
	select {
	case reply = <-pr.Replies[packet.ID]:
		pr.Logger.Debug("Peer", "%02X: Got Reply %02X for packet ID %02X", pr.ServerNetworkNode.ID, reply.ID, packet.ID)
		ticker.Stop()
		return reply, nil
	case <-ticker.C:
		pr.Logger.Warn("Peer", "%02X: Reply Timeout for packet ID %02X", pr.ServerNetworkNode.ID, packet.ID)
		return nil, errors.New("Reply Timeout")
	}
}

func (pr *Peer) handleKVStorePacket(packet *packets.Packet) {
	kvpacket := packet.Payload.(packets.KVStorePacket)
	switch kvpacket.Command {
	case packets.CMD_KVSTORE_SET:
		pr.handleKVStoreSet(&kvpacket, packet)
	case packets.CMD_KVSTORE_GET:
		pr.handleKVStoreGet(&kvpacket, packet)
	case packets.CMD_KVSTORE_DELETE:
		pr.handleKVStoreDelete(&kvpacket, packet)
	default:
		pr.Logger.Error("Peer", "KVStorePacket: Unknown Command %d", packet.Command)
	}
}

func (pr *Peer) handleKVStoreSet(packet *packets.KVStorePacket, request *packets.Packet) {
	pr.Logger.Debug("Peer", "%02X: KVStoreSet: %s = %s", pr.ServerNetworkNode.ID, packet.Key, packet.Data)
	pr.Server.KVStore.Set(
		packet.Key,
		packet.Data,
		packet.Flags,
		packet.ExpiresAt)

	response := packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
	pr.Logger.Debug("Peer", "%02X: KVStoreSet: %s Acknowledge, replying", pr.ServerNetworkNode.ID, packet.Key)
	pr.SendPacket(response)
}

func (pr *Peer) handleKVStoreGet(packet *packets.KVStorePacket, request *packets.Packet) {
	pr.Logger.Debug("Peer", "%02X: KVStoreGet: %s", pr.ServerNetworkNode.ID, packet.Key)
	value, flags, found := pr.Server.KVStore.Get(packet.Key)

	var response *packets.Packet

	if found {
		payload := packets.KVStorePacket{
			Command: packets.CMD_KVSTORE_GET,
			Key:     packet.Key,
			Data:    value,
			Flags:   flags,
		}
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, payload)
		pr.Logger.Debug("Peer", "%02X: KVStoreGet: %s = %s, replying", pr.ServerNetworkNode.ID, packet.Key, value)
	} else {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
		pr.Logger.Debug("Peer", "%02X: KVStoreGet: %s Not found, replying", pr.ServerNetworkNode.ID, packet.Key)
	}

	pr.SendPacket(response)
}

func (pr *Peer) handleKVStoreDelete(packet *packets.KVStorePacket, request *packets.Packet) {
	pr.Logger.Debug("Peer", "%02X: KVStoreDelete: %s", pr.ServerNetworkNode.ID, packet.Key)
	found := pr.Server.KVStore.Delete(packet.Key)

	var response *packets.Packet

	if found {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
		pr.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Deleted, replying", pr.ServerNetworkNode.ID, packet.Key)
	} else {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
		pr.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Not found, replying", pr.ServerNetworkNode.ID, packet.Key)
	}

	pr.SendPacket(response)
}
