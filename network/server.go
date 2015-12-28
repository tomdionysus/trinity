package network

import (
  "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/md5"
	"net"
	"github.com/tomdionysus/trinity/util"
	"github.com/tomdionysus/trinity/kvstore"
	"github.com/tomdionysus/trinity/packets"
	"github.com/tomdionysus/trinity/consistenthash"
	"fmt"
	"encoding/gob"
	"errors"
	"time"
)

const (
	CmdStop = iota
	StatusStopped = iota
)

type TLSServer struct {
	ServerNode *consistenthash.ServerNode

	CACertificate *tls.Certificate
	Certificate *tls.Certificate
	Logger *util.Logger
	ControlChannel chan(int)
	StatusChannel chan(int)

	CAPool *CAPool
	KVStore *kvstore.KVStore

	SessionCache tls.ClientSessionCache
	Connections map[[16]byte]*Peer

	Listener net.Listener
}

func NewTLSServer(logger *util.Logger, caPool *CAPool, kvStore *kvstore.KVStore, hostname string) *TLSServer {
	inst := &TLSServer{
		ServerNode: consistenthash.NewServerNode(hostname),
		Logger: logger,
		ControlChannel: make(chan(int)),
		StatusChannel: make(chan(int)),
		Connections: map[[16]byte]*Peer{},
		SessionCache: tls.NewLRUClientSessionCache(1024),
		KVStore: kvStore,
		CAPool: caPool,
	}
	inst.Logger.Debug("Server","Trinity Node ID %02X", inst.ServerNode.ID)
	return inst 
}

func (me *TLSServer) LoadPEMCert(certFile string, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		me.Certificate = &cert
		pcert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			me.Certificate.Leaf = pcert
		}
	}
	return err
}

func (me *TLSServer) ConnectTo(remoteAddr string) error {
	if me.Listener.Addr().String() == remoteAddr {
		er := "Cannot Connect to self"
		me.Logger.Error("Server",er) 
		return errors.New(er)
	}
	peer := NewPeer(me.Logger, me, remoteAddr)
	err := peer.Connect()
	if err != nil {
		me.Logger.Error("Server","Cannot Connect to Node %s: %s", remoteAddr, err.Error()) 
		return err
	}
	peer.Start()
	return nil
}

func (me *TLSServer) Listen(port uint16) error {
	config := tls.Config{
		ClientCAs: me.CAPool.Pool,
		ClientSessionCache: me.SessionCache,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{*me.Certificate},
		CipherSuites: []uint16{ 0x0035, 0xc030, 0xc02c },
	}

	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), &config)
	if err != nil {
		me.Logger.Error("Server","Cannot listen on port %d", port)
		return err
	}
	me.Listener = listener

	me.Logger.Info("Server","Listening on port %d", port)

	go me.server_loop()

	return nil
}

func (me *TLSServer) Stop() {
	me.ControlChannel <- CmdStop
}

func (me *TLSServer) IsConnectedTo(remoteAddr string) bool {
	for _, peer := range me.Connections {
		if remoteAddr == peer.ServerNetworkNode.HostAddr { return true }
		if remoteAddr == peer.Connection.RemoteAddr().String() { return true }
	}
	return false
}

func (me *TLSServer) NotifyNewPeer(newPeer *Peer) {
	me.Logger.Info("Server","Notifying Existing Peers of new Peer %02X (%s)", newPeer.ServerNetworkNode.ID, newPeer.ServerNetworkNode.HostAddr)
	for _, peer := range me.Connections { 
		if peer.ServerNetworkNode.ID != newPeer.ServerNetworkNode.ID {
			packet := packets.NewPacket(packets.CMD_PEERLIST, []string{ newPeer.ServerNetworkNode.HostAddr })
			peer.SendPacket(packet)
		}
	}
}

// Distributed Key Value store methods

func (me *TLSServer) SetKey(key string, value []byte, flags int16, expiry *time.Time) {
	keymd5 := getMD5(key)
	node := me.ServerNode.GetNodeFor(keymd5)
	if node.ID == me.ServerNode.ID {
		me.Logger.Debug("Server","SetKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
		// Local set.
		me.KVStore.Set(key, value, flags, expiry)
	} else {
		me.Logger.Debug("Server","SetKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)
		// Remote Set.
		payload := packets.KVStorePacket{
			Command: packets.CMD_KVSTORE_SET,
			Key: key,
			KeyHash: keymd5,
			Data: value,
			ExpiresAt: expiry,
			Flags: flags,
			TargetID: node.ID,
		}
		packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
		me.Connections[node.ID].SendPacketWaitReply(packet, 0)
	}
}

func (me *TLSServer) GetKey(key string) ([]byte, int16, bool) {
	keymd5 := getMD5(key)
	node := me.ServerNode.GetNodeFor(keymd5)
	if node.ID == me.ServerNode.ID {
		me.Logger.Debug("Server","GetKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
		// Local set.
		return me.KVStore.Get(key)
	} else {
		me.Logger.Debug("Server","GetKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)
		// Remote Set.
		payload := packets.KVStorePacket{
			Command: packets.CMD_KVSTORE_GET,
			Key: key,
			KeyHash: keymd5,
			TargetID: node.ID,
		}
		packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
		reply, err := me.Connections[node.ID].SendPacketWaitReply(packet, 0)
		
		// Process reply or timeout
		if err==nil {
			switch reply.Command {
			case packets.CMD_KVSTORE_ACK:
				kvpacket := reply.Payload.(packets.KVStorePacket)
				me.Logger.Debug("Server","GetKey: Reply from Remote %s = %s", key, kvpacket.Data)
				return kvpacket.Data, kvpacket.Flags, true
			case packets.CMD_KVSTORE_NOT_FOUND:
				me.Logger.Debug("Server","GetKey: Reply from Remote %s Not Found", key)
				return []byte{}, 0, false
			default:
				me.Logger.Warn("Server","GetKey: Unknown Reply Command %d", reply.Command)
			}
		} else {
			me.Logger.Warn("Server","GetKey: Reply Timeout")
			// TODO: Timeout
		}
	}
	return []byte{}, 0, false
}

func (me *TLSServer) DeleteKey(key string) bool {
	keymd5 := getMD5(key)
	node := me.ServerNode.GetNodeFor(keymd5)
	if node.ID == me.ServerNode.ID {
		me.Logger.Debug("Server","DeleteKey: Peer for key %02X -> %02X (Local)", keymd5, node.ID)
		// Local set.
		return me.KVStore.Delete(key)
	} else {
		me.Logger.Debug("Server","DeleteKey: Peer for key %02X -> %02X (Remote)", keymd5, node.ID)
		// Remote Set.
		payload := packets.KVStorePacket{
			Command: packets.CMD_KVSTORE_DELETE,
			Key: key,
			KeyHash: keymd5,
			TargetID: node.ID,
		}
		packet := packets.NewPacket(packets.CMD_KVSTORE, payload)
		reply, err := me.Connections[node.ID].SendPacketWaitReply(packet, 0)
		
		// Process reply or timeout
		if err==nil {
			switch reply.Command {
			case packets.CMD_KVSTORE_ACK:
				me.Logger.Debug("Server","DeleteKey: Reply from Remote %s Deleted", key)
				return true
			case packets.CMD_KVSTORE_NOT_FOUND:
				me.Logger.Debug("Server","DeleteKey: Reply from Remote %s Not Found", key)
				return false
			default:
				me.Logger.Warn("Server","DeleteKey: Unknown Reply Command %d", reply.Command)
			}
		} else {
			me.Logger.Warn("Server","DeleteKey: Reply Timeout")
			// TODO: Timeout
		}
	}
	return false
}

func getMD5(keystring string) consistenthash.Key {
	return consistenthash.Key(md5.Sum([]byte(keystring)))
}

// func (me *TLSServer) Broadcast()

func (me *TLSServer) server_loop() {

	// Connection Acceptor Loop
	go func() {
		for {
			conn, err := me.Listener.Accept()
			if err != nil {
				me.Logger.Error("Server","Cannot Accept connection from %s: %s", conn.RemoteAddr(), err.Error())
				break
			}
  		me.Logger.Debug("Server", "Incoming Connection From %s", conn.RemoteAddr())
			peer := NewConnectingPeer(me.Logger, me, conn.(*tls.Conn))
			peer.Start()
		}
	}()

	// Control / Stop loop
	for {
		select {
			case cmd := <- me.ControlChannel:
				if cmd == CmdStop {
					me.Logger.Debug("Server","Stop Recieved, Shutting Down")
					goto end
				}
		}
	}
	
	end:

	for _, Peer := range me.Connections {
		me.Logger.Debug("Server","Closing Connection %s", Peer.Connection.RemoteAddr())
		Peer.Disconnect()
	}

	me.Logger.Info("Server","Stopped")
	me.StatusChannel <- StatusStopped
}

func init() {
	gob.Register(consistenthash.ServerNetworkNode{})
}

