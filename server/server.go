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

type TLSServer struct {
	Certificate *tls.Certificate
	Logger *util.Logger

	listener net.Listener
	ControlChannel chan(int)
	StatusChannel chan(int)
}

func NewTLSServer(logger *util.Logger) *TLSServer {
	return &TLSServer{
		Logger: logger,
		ControlChannel: make(chan(int)),
		StatusChannel: make(chan(int)),
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
		ClientAuth: tls.NoClientCert,
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

func (me *TLSServer) Stop() {
	me.ControlChannel <- CmdStop
}

func (me *TLSServer) server_loop() {
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
	
	me.Logger.Info("Server","Stopped")
	me.StatusChannel <- StatusStopped
}
