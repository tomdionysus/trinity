package server

import (
  "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net"
	"github.com/tomdionysus/trinity/util"
	"fmt"
	"io/ioutil"
	"errors"
)

const (
	CmdStop = iota
	StatusStopped = iota
)

type Client struct {
	Connection net.Conn
}

type TLSServer struct {
	CACertificate *tls.Certificate
	Certificate *tls.Certificate
	Logger *util.Logger
	ControlChannel chan(int)
	StatusChannel chan(int)

	CAPool *x509.CertPool

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
		CAPool: x509.NewCertPool(),
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

func (me *TLSServer) LoadPEMCA(certFile string) error {
	cabytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		me.Logger.Error("Server","Cannot Load CA File '%s': %s", certFile, err.Error())
		return err
	}

	if !me.CAPool.AppendCertsFromPEM(cabytes) {
		me.Logger.Error("Server","Cannot Parse PEM CA File '%s'", certFile)
		return errors.New("Cannot Parse CA File")
	}

	return nil
}

func (me *TLSServer) Listen(port uint16) error {
	config := tls.Config{
		ClientCAs: me.CAPool,
		InsecureSkipVerify: true,
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

func (me *TLSServer) ConnectTo(url string) {
	conn, err := tls.Dial("tcp", url, &tls.Config{
		RootCAs: me.CAPool,
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{*me.Certificate},
	})
	if err != nil {
		me.Logger.Error("Server","Cannot connect to %s: %s", url, err.Error())
		return
	}
	state := conn.ConnectionState()
	if len(state.PeerCertificates)==0 {
		me.Logger.Error("Server","Cannot connect to %s: Peer has no certificates", url)
		conn.Close()
		return
	}
	if !state.HandshakeComplete {
		me.Logger.Error("Server","Cannot connect to %s: Handshake Not Complete", url)
		conn.Close()
		return
	}
	sub := state.PeerCertificates[0].Subject.CommonName
	me.Logger.Info("Server","Connected To %s (%s) [%s]", conn.RemoteAddr(), sub, Ciphers[state.CipherSuite])
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
	sub := state.PeerCertificates[0].Subject.CommonName
	me.Logger.Info("Server","Connection Accepted from %s (%s) [%s]",client.Connection.RemoteAddr(), sub, Ciphers[tlscon.ConnectionState().CipherSuite])
}

var Ciphers map[uint16]string = map[uint16]string{
	0x0005: "TLS_RSA_WITH_RC4_128_SHA",
	0x000a: "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
	0x002f: "TLS_RSA_WITH_AES_128_CBC_SHA",
	0x0035: "TLS_RSA_WITH_AES_256_CBC_SHA",
	0xc007: "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
	0xc009: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	0xc00a: "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	0xc011: "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
	0xc012: "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
	0xc013: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	0xc014: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	0xc02f: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	0xc02b: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	0xc030: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	0xc02c: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
}

