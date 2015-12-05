package server

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

type Client struct {
	Connection net.Conn
}

type TLSServer struct {
	Certificate *tls.Certificate
	Logger *util.Logger
	ControlChannel chan(int)
	StatusChannel chan(int)

	SessionCache tls.ClientSessionCache
	Connections map[string]Client

	listener net.Listener
}

func NewTLSServer(logger *util.Logger) *TLSServer {
	return &TLSServer{
		Logger: logger,
		ControlChannel: make(chan(int)),
		StatusChannel: make(chan(int)),
		Connections: map[string]Client{},
		SessionCache: tls.NewLRUClientSessionCache(64),
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
		ClientSessionCache: me.SessionCache,
		ClientAuth: tls.RequireAnyClientCert,
		Certificates: []tls.Certificate{*me.Certificate},
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

func (me *TLSServer) ConnectTo(url string) {
	conn, err := tls.Dial("tcp", url, &tls.Config{
		Certificates: []tls.Certificate{*me.Certificate},
	})
	if err != nil {
		me.Logger.Error("Server","Cannot connect to %s",url)
		return
	}
	me.Logger.Info("Server","Connected To %s (%s)", url, conn.RemoteAddr())
}

func (me *TLSServer) Stop() {
	me.ControlChannel <- CmdStop
}

func (me *TLSServer) server_loop() {

	go func() {
		for {
			conn, err := me.listener.Accept()
			if err != nil {
				me.Logger.Error("Server","Cannot Accept connection from %s: %s", conn.RemoteAddr(), err.Error())
				break
			}
			client := &Client{
				Connection: conn,
			}
			go me.handle_client(client)
		}
	}()

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

	for _, client := range me.Connections {
		me.Logger.Debug("Server","Closing Connection %s", client.Connection.RemoteAddr())
		client.Connection.Close()
	}

	me.Logger.Info("Server","Stopped")
	me.StatusChannel <- StatusStopped
}

func (me *TLSServer) handle_client(client *Client) {
	me.Logger.Info("Server","Connection accepted from %s", client.Connection.RemoteAddr())
	tlscon, ok := client.Connection.(*tls.Conn)
	if !ok {
		me.Logger.Error("Server","Cannot Cast Connection to TLS Connection")
		return
	}

	err := tlscon.Handshake()
	if err!=nil {
		me.Logger.Info("Server","Client TLS Handshake failed, closing: %s",err.Error())
		tlscon.Close()
		return
	}
	state := tlscon.ConnectionState()
	if len(state.PeerCertificates)==0 {
		me.Logger.Info("Server","Client has no certificates, closing")
		tlscon.Close()
		return
	}
	sub := state.PeerCertificates[0].Subject
	me.Logger.Info("Server","Connection Subject %s",sub)
}
