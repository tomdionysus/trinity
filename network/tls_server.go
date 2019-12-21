package network

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	ch "github.com/tomdionysus/consistenthash"
	"github.com/tomdionysus/trinity/kvstore"
	"github.com/tomdionysus/trinity/packets"
	"github.com/tomdionysus/trinity/util"
)

// Commands
const (
	CommandStop = iota
)

// Status
const (
	StatusStopped = iota
)

// TLSServer represents the running Trinity instance, and is the highest level type in the stack.
type TLSServer struct {
	ServerNode *ch.ServerNode

	CACertificate  *tls.Certificate
	Certificate    *tls.Certificate
	Logger         *util.Logger
	ControlChannel chan (int)
	StatusChannel  chan (int)

	CAPool  *CAPool
	KVStore *kvstore.KVStore

	SessionCache tls.ClientSessionCache

	connections      map[ch.NodeId]*Peer
	connectionsMutex sync.Mutex

	Listener net.Listener
}

// NewTLSServer creates and returns a new TLSServer with the given logger, CA Pool, KV Store and host name
func NewTLSServer(logger *util.Logger, caPool *CAPool, kvStore *kvstore.KVStore, hostname string) *TLSServer {
	inst := &TLSServer{
		ServerNode:     ch.NewServerNode(hostname),
		Logger:         logger,
		ControlChannel: make(chan (int)),
		StatusChannel:  make(chan (int)),
		SessionCache:   tls.NewLRUClientSessionCache(1024),
		KVStore:        kvStore,
		CAPool:         caPool,

		connections: map[ch.NodeId]*Peer{},
	}
	return inst
}

// LoadPEMCert loads the given PEM file(s) as this instance's certificate. certFile and keyFile may be the same file.
func (svr *TLSServer) LoadPEMCert(certFile string, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		svr.Certificate = &cert
		pcert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			svr.Certificate.Leaf = pcert
		}
	}
	return err
}

// ConnectTo attempts to connect to another Trinity instance and join its cluster.
func (svr *TLSServer) ConnectTo(remoteAddr string) error {
	if svr.Listener.Addr().String() == remoteAddr {
		er := "Cannot Connect to self"
		svr.Logger.Error("Server", er)
		return errors.New(er)
	}
	peer := NewPeer(svr.Logger, svr, remoteAddr)
	err := peer.Connect()
	if err != nil {
		svr.Logger.Error("Server", "Cannot Connect to Node %s: %s", remoteAddr, err.Error())
		return err
	}
	peer.Start()
	return nil
}

// Listen begins listening for other Trinity instances connecting on the given port.
func (svr *TLSServer) Listen(port uint16) error {
	config := tls.Config{
		ClientCAs:          svr.CAPool.Pool,
		ClientSessionCache: svr.SessionCache,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		Certificates:       []tls.Certificate{*svr.Certificate},
		CipherSuites:       []uint16{0x0035, 0xc030, 0xc02c},
	}

	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), &config)
	if err != nil {
		svr.Logger.Error("Server", "Cannot listen on port %d", port)
		return err
	}
	svr.Listener = listener

	svr.Logger.Info("Server", "Listening on port %d", port)

	go svr.server_loop()

	return nil
}

// Stop causes the instance to stop listening, close all connections, and shut down.
func (svr *TLSServer) Stop() {
	svr.ControlChannel <- CommandStop
}

// Connections Concurrent Map

// ConnectionSet assigns the given Instance ID to the given peer
func (svr *TLSServer) ConnectionSet(id ch.NodeId, peer *Peer) {
	svr.connectionsMutex.Lock()
	svr.connections[id] = peer
	svr.connectionsMutex.Unlock()
}

// ConnectionGet returns the peer for the given ID, and whether that ID was found.
func (svr *TLSServer) ConnectionGet(id ch.NodeId) (*Peer, bool) {
	svr.connectionsMutex.Lock()
	peer, found := svr.connections[id]
	svr.connectionsMutex.Unlock()
	return peer, found
}

// ConnectionClear clears the peer for the given Instance ID.
func (svr *TLSServer) ConnectionClear(id ch.NodeId) {
	svr.connectionsMutex.Lock()
	delete(svr.connections, id)
	svr.connectionsMutex.Unlock()
}

// Connections returns a current copy of all connections.
func (svr *TLSServer) Connections() map[ch.NodeId]*Peer {
	svr.connectionsMutex.Lock()
	cpy := map[ch.NodeId]*Peer{}

	for k, v := range svr.connections {
		cpy[k] = v
	}

	svr.connectionsMutex.Unlock()
	return cpy
}

// NotifyAllPeers sends CMD_PEERLIST to all connected peers with a list of all the peers we know about.
func (svr *TLSServer) NotifyAllPeers() {
	svr.Logger.Debug("Server", "Notifying All Peers")

	currentConns := svr.Connections()

	for id, peer := range currentConns {
		payload := packets.PeerListPacket{}
		for listId, listPeer := range currentConns {
			if listId != id {
				payload[listId] = listPeer.ServerNetworkNode.HostAddr
			}
		}
		peer.SendPacket(packets.NewPacket(packets.CMD_PEERLIST, payload))
	}
}

// SetKey sets the given key to the given value in the cluster.
func (svr *TLSServer) SetKey(key string, value []byte, flags int16, expiry *time.Time) {
	keymd5 := ch.NewMD5Key(key)
	nodes := svr.ServerNode.GetNodesFor(keymd5, 3)
	svr.Logger.Debug("Server", "SetKey: %d peers for key %02X", len(nodes), keymd5)
	for _, node := range nodes {
		if node.ID == svr.ServerNode.ID {
			svr.Logger.Debug("Server", "SetKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
			// Local set.
			svr.KVStore.Set(key, value, flags, expiry)
		} else {
			svr.Logger.Debug("Server", "SetKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)

			peer, _ := svr.ConnectionGet(node.ID)
			if peer.State != PeerStateConnected {
				svr.Logger.Warn("Server", "SetKey: Peer for key %02X -> %02X (Remote) Unavailable", keymd5, node.ID)
				continue
			}
			// Remote Set.
			payload := packets.KVStorePacket{
				Command:   packets.CMD_KVSTORE_SET,
				Key:       key,
				KeyHash:   keymd5,
				Data:      value,
				ExpiresAt: expiry,
				Flags:     flags,
				TargetID:  node.ID,
			}
			packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
			peer.SendPacketWaitReply(packet, 5*time.Second)
		}
	}
}

// GetKey returns a value for the given key in the cluster, and if that key was found
func (svr *TLSServer) GetKey(key string) ([]byte, int16, bool) {
	keymd5 := ch.NewMD5Key(key)
	nodes := svr.ServerNode.GetNodesFor(keymd5, 3)
	for _, node := range nodes {
		if node.ID == svr.ServerNode.ID {
			svr.Logger.Debug("Server", "GetKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
			// Local set.
			return svr.KVStore.Get(key)
		} else {
			svr.Logger.Debug("Server", "GetKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)

			peer, _ := svr.ConnectionGet(node.ID)
			if peer.State != PeerStateConnected {
				svr.Logger.Warn("Server", "GetKey: Peer for key %02X -> %02X (Remote) Unavailable", keymd5, node.ID)
				continue
			}

			// Remote Set.
			payload := packets.KVStorePacket{
				Command:  packets.CMD_KVSTORE_GET,
				Key:      key,
				KeyHash:  keymd5,
				TargetID: node.ID,
			}
			packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
			reply, err := peer.SendPacketWaitReply(packet, 5*time.Second)

			// Process reply or timeout
			if err == nil {
				switch reply.Command {
				case packets.CMD_KVSTORE_ACK:
					kvpacket := reply.Payload.(packets.KVStorePacket)
					svr.Logger.Debug("Server", "GetKey: Reply from Remote %s = %s", key, kvpacket.Data)
					return kvpacket.Data, kvpacket.Flags, true
				case packets.CMD_KVSTORE_NOT_FOUND:
					svr.Logger.Debug("Server", "GetKey: Reply from Remote %s Not Found", key)
					return []byte{}, 0, false
				default:
					svr.Logger.Warn("Server", "GetKey: Unknown Reply Command %d", reply.Command)
				}
			} else {
				svr.Logger.Warn("Server", "GetKey: Reply Timeout")
			}
		}
	}
	return []byte{}, 0, false
}

// IsSet return if a key is set
func (svr *TLSServer) IsSet(key string) bool {
	keymd5 := ch.NewMD5Key(key)
	nodes := svr.ServerNode.GetNodesFor(keymd5, 3)
	for _, node := range nodes {
		if node.ID == svr.ServerNode.ID {
			svr.Logger.Debug("Server", "GetKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
			// Local set.
			return svr.KVStore.IsSet(key)
		} else {
			svr.Logger.Debug("Server", "GetKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)

			peer, _ := svr.ConnectionGet(node.ID)
			if peer.State != PeerStateConnected {
				svr.Logger.Warn("Server", "GetKey: Peer for key %02X -> %02X (Remote) Unavailable", keymd5, node.ID)
				continue
			}

			// Remote Set.
			payload := packets.KVStorePacket{
				Command:  packets.CMD_KVSTORE_IS_SET,
				Key:      key,
				KeyHash:  keymd5,
				TargetID: node.ID,
			}
			packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
			reply, err := peer.SendPacketWaitReply(packet, 5*time.Second)

			// Process reply or timeout
			if err == nil {
				switch reply.Command {
				case packets.CMD_KVSTORE_ACK:
					svr.Logger.Debug("Server", "GetKey: Reply from Remote %s", key)
					return true
				case packets.CMD_KVSTORE_NOT_FOUND:
					svr.Logger.Debug("Server", "GetKey: Reply from Remote %s Not Found", key)
					return false
				default:
					svr.Logger.Warn("Server", "GetKey: Unknown Reply Command %d", reply.Command)
				}
			} else {
				svr.Logger.Warn("Server", "GetKey: Reply Timeout")
			}
		}
	}
	return false
}

// DeleteKey clears the given key in the cluster.
func (svr *TLSServer) DeleteKey(key string) bool {
	keymd5 := ch.NewMD5Key(key)
	node := svr.ServerNode.GetNodeFor(keymd5)
	if node.ID == svr.ServerNode.ID {
		svr.Logger.Debug("Server", "DeleteKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
		// Local set.
		return svr.KVStore.Delete(key)
	} else {
		svr.Logger.Debug("Server", "DeleteKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)
		// Remote Set.
		payload := packets.KVStorePacket{
			Command:  packets.CMD_KVSTORE_DELETE,
			Key:      key,
			KeyHash:  keymd5,
			TargetID: node.ID,
		}

		packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
		peer, _ := svr.ConnectionGet(node.ID)
		reply, err := peer.SendPacketWaitReply(packet, 5*time.Second)

		// Process reply or timeout
		if err == nil {
			switch reply.Command {
			case packets.CMD_KVSTORE_ACK:
				svr.Logger.Debug("Server", "DeleteKey: Reply from Remote %s Deleted", key)
				return true
			case packets.CMD_KVSTORE_NOT_FOUND:
				svr.Logger.Debug("Server", "DeleteKey: Reply from Remote %s Not Found", key)
				return false
			default:
				svr.Logger.Warn("Server", "DeleteKey: Unknown Reply Command %d", reply.Command)
			}
		} else {
			svr.Logger.Warn("Server", "DeleteKey: Reply Timeout")
		}
	}
	return false
}

// server_loop starts the main server runloop, accepting connections. The peers have individual runloops which handle
// incoming traffic.
func (svr *TLSServer) server_loop() {

	go func() {
		for {
			conn, err := svr.Listener.Accept()
			if err != nil {
				svr.Logger.Error("Server", "Cannot Accept connection (%s)", err.Error())
				break
			}
			svr.Logger.Debug("Server", "Incoming Connection From %s", conn.RemoteAddr())
			peer := NewConnectingPeer(svr.Logger, svr, conn.(*tls.Conn))
			peer.Start()
		}
	}()

	// Control / Stop loop
	for {
		select {
		case cmd := <-svr.ControlChannel:
			switch cmd {
			case CommandStop:
				svr.Logger.Debug("Server", "Stop Received, Shutting Down")
				goto end
			}
		}
	}

end:

	svr.Logger.Debug("Server", "Closing Peer Connections")
	for _, peer := range svr.Connections() {
		peer.Disconnect()
	}

	svr.Logger.Info("Server", "Stopped")
	svr.StatusChannel <- StatusStopped
}

// init registers ServerNetworkNode with GOB for packet transfer
func init() {
	gob.Register(ch.ServerNetworkNode{})
}
