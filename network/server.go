package network

import (
  "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net"
	"github.com/tomdionysus/trinity/util"
	"github.com/tomdionysus/trinity/kvstore"
	"github.com/tomdionysus/trinity/packets"
	"github.com/tomdionysus/trinity/consistenthash"
	"fmt"
	"encoding/gob"
	"errors"
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
	return &TLSServer{
		ServerNode: consistenthash.NewServerNode(hostname),
		Logger: logger,
		ControlChannel: make(chan(int)),
		StatusChannel: make(chan(int)),
		Connections: map[[16]byte]*Peer{},
		SessionCache: tls.NewLRUClientSessionCache(1024),
		KVStore: kvStore,
		CAPool: caPool,
	}
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

