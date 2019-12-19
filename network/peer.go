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

// Peer States
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

// Peer is a representation of a remote trinity instance.
type Peer struct {
	Logger *util.Logger

	// Server is the TLSServer that the peer is connected to (i.e. the TLSServer in this process)
	Server *TLSServer

	// Incoming is true if the peer connected to this server, false if we connected to it
	Incoming bool

	// Address is the hostname:port to connect to, or the hostname:port being connected
	Address string

	// State is the State of the peer
	State uint

	// Connecton is the underlying TLS secured connection
	Connection *tls.Conn

	// HeartbeatTicker is the ticker used to generate heartbeat packets (every 1s)
	HeartbeatTicker *time.Ticker

	// Writer is the stream for sending to the peer
	Writer *gob.Encoder
	// Reader is the stream for reading from the peer
	Reader *gob.Decoder

	// LastHeartbeat is when the last heartbeat packet was received
	LastHeartbeat time.Time

	// ServerNetworkNode is the consisten hash node associated with the peer
	ServerNetworkNode *consistenthash.ServerNetworkNode

	// Replies contains the current outstanding requests to the peer
	Replies map[packets.PacketId]chan (*packets.Packet)
}

// NewPeer returns a new Peer with the specified logger, server and address
func NewPeer(logger *util.Logger, server *TLSServer, address string) *Peer {
	inst := &Peer{
		Logger:            logger,
		Address:           address,
		State:             PeerStateDisconnected,
		Server:            server,
		LastHeartbeat:     time.Now(),
		ServerNetworkNode: nil,
		Replies:           map[packets.PacketId]chan (*packets.Packet){},
	}
	return inst
}

// NewConnectingPeer returns a new Peer with the specified logger, server, and connection.
func NewConnectingPeer(logger *util.Logger, server *TLSServer, connection *tls.Conn) *Peer {
	inst := NewPeer(logger, server, connection.RemoteAddr().String())
	inst.Connection = connection
	inst.State = PeerStateHandshake
	inst.Incoming = true
	return inst
}

// Connect attempts to connect to the trinity instance at the configred Address
func (peer *Peer) Connect() error {
	peer.Incoming = false
	peer.State = PeerStateConnecting
	conn, err := tls.Dial("tcp", peer.Address, &tls.Config{
		RootCAs:      peer.Server.CAPool.Pool,
		Certificates: []tls.Certificate{*peer.Server.Certificate},
	})
	if err != nil {
		peer.Logger.Error("Peer", "Cannot connect to %s: %s", peer.Address, err.Error())
		return err
	}
	peer.Connection = conn
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		peer.Logger.Error("Peer", "Cannot connect to %s: Peer has no certificates", peer.Address)
		peer.Disconnect()
		return errors.New("Peer has no certificates")
	}
	peer.State = PeerStateHandshake
	return nil
}

// Disconnect disconnects the remote trinity instance and removes the peer from the TLSServer connections
func (peer *Peer) Disconnect() {
	if peer.ServerNetworkNode != nil && peer.State != PeerStateDisconnected {
		peer.State = PeerStateDisconnected
		peer.Server.ServerNode.DeregisterNode(peer.ServerNetworkNode)
		if peer.HeartbeatTicker != nil {
			peer.HeartbeatTicker.Stop()
		}
		if peer.Connection != nil {
			peer.Connection.Close()
		}
		peer.Logger.Info("Peer", "%02X: Disconnected", peer.ServerNetworkNode.ID)
		peer.Server.ConnectionClear(peer.ServerNetworkNode.ID)
	}
}

// Start processes the TLS handshake and registration protocol once connected
func (peer *Peer) Start() error {
	if peer.State != PeerStateHandshake {
		peer.Logger.Error("Peer", "Cannot Start Peer, Handshake not ready")
		return errors.New("Handshake not ready")
	}
	err := peer.Connection.Handshake()
	if err != nil {
		peer.Logger.Error("Peer", "Peer TLS Handshake failed, disconnecting: %s", err.Error())
		peer.Disconnect()
		return errors.New("TLS Handshake Failed")
	}
	state := peer.Connection.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		peer.Logger.Error("Peer", "Peer has no certificates, disconnecting")
		peer.Disconnect()
		return errors.New("Peer sent no certificates")
	}
	sub := state.PeerCertificates[0].Subject.CommonName

	if peer.Incoming {
		peer.Logger.Info("Peer", "Outgoing Connection to %s (%s) [%s]", peer.Connection.RemoteAddr(), sub, Ciphers[peer.Connection.ConnectionState().CipherSuite])
	} else {
		peer.Logger.Info("Peer", "Incoming Connection from %s (%s) [%s]", peer.Connection.RemoteAddr(), sub, Ciphers[peer.Connection.ConnectionState().CipherSuite])
	}

	peer.Reader = gob.NewDecoder(peer.Connection)
	peer.Writer = gob.NewEncoder(peer.Connection)

	go peer.heartbeat()

	peer.SendDistribution()

	go peer.process()

	return nil
}

// heartbeat pings the Peer every second.
func (peer *Peer) heartbeat() {
	peer.HeartbeatTicker = time.NewTicker(time.Second)

	for {
		<-peer.HeartbeatTicker.C

		// Check For Defib
		if time.Now().After(peer.LastHeartbeat.Add(5 * time.Second)) {
			peer.Logger.Warn("Peer", "%02X: Peer Defib (no response for >5 seconds)", peer.ServerNetworkNode.ID)
			peer.State = PeerStateDefib
		}

		switch peer.State {
		case PeerStateConnected:
			err := peer.SendPacket(packets.NewPacket(packets.CMD_HEARTBEAT, nil))
			if err != nil {
				peer.Logger.Error("Peer", "%02X: Error Sending Heartbeat, disconnecting", peer.ServerNetworkNode.ID)
				peer.Disconnect()
				return
			}
		case PeerStateDefib:
			if time.Now().After(peer.LastHeartbeat.Add(10 * time.Second)) {
				peer.Logger.Warn("Peer", "%02X: Peer DOA (Defib for 5 seconds), disconnecting", peer.ServerNetworkNode.ID)
				peer.Disconnect()
				return
			}
		}
	}
}

// process continually reads from the Pere input stream and processes packet commands.
func (peer *Peer) process() {

	var packet packets.Packet

	for {

		// Read Command
		err := peer.Reader.Decode(&packet)
		if err != nil {
			if err.Error() == "EOF" {
				peer.Logger.Debug("Peer", "%02X: Peer Closed Connection", peer.ServerNetworkNode.ID)
			} else {
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					peer.Logger.Debug("Peer", "%02X: Read After This Node Closed Connection", peer.ServerNetworkNode.ID)
				} else {
					peer.Logger.Error("Peer", "%02X: Error Reading: %s", peer.ServerNetworkNode.ID, err.Error())
				}
			}
			goto end
		}

		switch packet.Command {

		// Packets in Connecting / Handshake

		case packets.CMD_HEARTBEAT:
			peer.LastHeartbeat = time.Now()

		case packets.CMD_DISTRIBUTION:
			peer.process_CMD_DISTRIBUTION(packet)

		// Packets in Connected

		case packets.CMD_PEERLIST:
			peer.process_CMD_PEERLIST(packet)

		case packets.CMD_KVSTORE:
			peer.Logger.Debug("Peer", "%02X: CMD_KVSTORE", peer.ServerNetworkNode.ID)
			peer.handleKVStorePacket(&packet)

		case packets.CMD_KVSTORE_ACK:
			peer.Logger.Debug("Peer", "%02X: CMD_KVSTORE_ACK", peer.ServerNetworkNode.ID)
			peer.handleReply(&packet)

		case packets.CMD_KVSTORE_NOT_FOUND:
			peer.Logger.Debug("Peer", "%02X: CMD_KVSTORE_NOT_FOUND", peer.ServerNetworkNode.ID)
			peer.handleReply(&packet)

		default:
			peer.Logger.Warn("Peer", "%02X: Unknown Packet Command %d", peer.ServerNetworkNode.ID, packet.Command)
		}

		if peer.State == PeerStateDisconnected {
			// As a result of the packet, the peer is now disconnected and we should not try to read from it further.
			goto end
		}
	}
end:

	peer.Disconnect()
}

func (peer *Peer) handleReply(packet *packets.Packet) {
	chn, found := peer.Replies[packet.RequestID]
	if found {
		delete(peer.Replies, packet.RequestID)
		chn <- packet
	} else {
		peer.Logger.Warn("Peer", "%02X: Unsolicited Reply to unknown packet %02X", peer.ServerNetworkNode.ID, packet.RequestID)
	}
}

func (peer *Peer) SendDistribution() error {
	packet := packets.NewPacket(packets.CMD_DISTRIBUTION, peer.Server.ServerNode.ServerNetworkNode)
	peer.SendPacket(packet)
	return nil
}

func (peer *Peer) SendPacket(packet *packets.Packet) error {
	err := peer.Writer.Encode(packet)
	if err != nil {
		peer.Logger.Error("Peer", "Error Writing: %s", err.Error())
	}
	return err
}

func (peer *Peer) SendPacketWaitReply(packet *packets.Packet, timeout time.Duration) (*packets.Packet, error) {
	if peer.State != PeerStateConnected {
		peer.Logger.Error("Peer", "%02X: Cannot send packet ID %02X, not PeerStateConnected", peer.ServerNetworkNode.ID, packet.ID)
		return nil, errors.New("Cannot send, state not PeerStateConnected")
	}

	peer.Replies[packet.ID] = make(chan (*packets.Packet))
	ticker := time.NewTicker(timeout)
	peer.SendPacket(packet)
	var reply *packets.Packet
	select {
	case reply = <-peer.Replies[packet.ID]:
		peer.Logger.Debug("Peer", "%02X: Got Reply %02X for packet ID %02X", peer.ServerNetworkNode.ID, reply.ID, packet.ID)
		ticker.Stop()
		return reply, nil
	case <-ticker.C:
		peer.Logger.Warn("Peer", "%02X: Reply Timeout for packet ID %02X", peer.ServerNetworkNode.ID, packet.ID)
		return nil, errors.New("Reply Timeout")
	}
}

func (peer *Peer) handleKVStorePacket(packet *packets.Packet) {
	kvpacket := packet.Payload.(packets.KVStorePacket)
	switch kvpacket.Command {
	case packets.CMD_KVSTORE_SET:
		peer.handleKVStoreSet(&kvpacket, packet)
	case packets.CMD_KVSTORE_GET:
		peer.handleKVStoreGet(&kvpacket, packet)
	case packets.CMD_KVSTORE_DELETE:
		peer.handleKVStoreDelete(&kvpacket, packet)
	default:
		peer.Logger.Error("Peer", "KVStorePacket: Unknown Command %d", packet.Command)
	}
}

func (peer *Peer) handleKVStoreSet(packet *packets.KVStorePacket, request *packets.Packet) {
	peer.Logger.Debug("Peer", "%02X: KVStoreSet: %s = %s", peer.ServerNetworkNode.ID, packet.Key, packet.Data)
	peer.Server.KVStore.Set(
		packet.Key,
		packet.Data,
		packet.Flags,
		packet.ExpiresAt)

	response := packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
	peer.Logger.Debug("Peer", "%02X: KVStoreSet: %s Acknowledge, replying", peer.ServerNetworkNode.ID, packet.Key)
	peer.SendPacket(response)
}

func (peer *Peer) handleKVStoreGet(packet *packets.KVStorePacket, request *packets.Packet) {
	peer.Logger.Debug("Peer", "%02X: KVStoreGet: %s", peer.ServerNetworkNode.ID, packet.Key)
	value, flags, found := peer.Server.KVStore.Get(packet.Key)

	var response *packets.Packet

	if found {
		payload := packets.KVStorePacket{
			Command: packets.CMD_KVSTORE_GET,
			Key:     packet.Key,
			Data:    value,
			Flags:   flags,
		}
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, payload)
		peer.Logger.Debug("Peer", "%02X: KVStoreGet: %s = %s, replying", peer.ServerNetworkNode.ID, packet.Key, value)
	} else {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
		peer.Logger.Debug("Peer", "%02X: KVStoreGet: %s Not found, replying", peer.ServerNetworkNode.ID, packet.Key)
	}

	peer.SendPacket(response)
}

func (peer *Peer) handleKVStoreDelete(packet *packets.KVStorePacket, request *packets.Packet) {
	peer.Logger.Debug("Peer", "%02X: KVStoreDelete: %s", peer.ServerNetworkNode.ID, packet.Key)
	found := peer.Server.KVStore.Delete(packet.Key)

	var response *packets.Packet

	if found {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
		peer.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Deleted, replying", peer.ServerNetworkNode.ID, packet.Key)
	} else {
		response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
		peer.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Not found, replying", peer.ServerNetworkNode.ID, packet.Key)
	}

	peer.SendPacket(response)
}
