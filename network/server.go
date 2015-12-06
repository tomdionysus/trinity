package network

import (
  "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net"
	"github.com/tomdionysus/trinity/util"
	"fmt"
)

const (
	CmdStop = iota
	StatusStopped = iota
)

type TLSServer struct {
	CACertificate *tls.Certificate
	Certificate *tls.Certificate
	Logger *util.Logger
	ControlChannel chan(int)
	StatusChannel chan(int)

	CAPool *CAPool

	SessionCache tls.ClientSessionCache
	Connections map[string]Peer

	listener net.Listener
}

func NewTLSServer(logger *util.Logger, caPool *CAPool) *TLSServer {
	return &TLSServer{
		Logger: logger,
		ControlChannel: make(chan(int)),
		StatusChannel: make(chan(int)),
		Connections: map[string]Peer{},
		SessionCache: tls.NewLRUClientSessionCache(64),
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
	me.listener = listener

	me.Logger.Info("Server","Listening on port %d", port)

	go me.server_loop()

	return nil
}

func (me *TLSServer) Stop() {
	me.ControlChannel <- CmdStop
}

func (me *TLSServer) server_loop() {

	// Connection Acceptor Loop
	go func() {
		for {
			conn, err := me.listener.Accept()
			if err != nil {
				me.Logger.Error("Server","Cannot Accept connection from %s: %s", conn.RemoteAddr(), err.Error())
				break
			}
  		me.Logger.Debug("Server", "Incoming Connection From %s", conn.RemoteAddr())
			Peer := NewConnectingPeer(me.Logger, me, conn.(*tls.Conn))
			Peer.Start()
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

