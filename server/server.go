package server

import (
	"crypto/tls"
	"crypto/x509"
)

type TLSServer struct {
	Certificate *tls.Certificate
}

func NewTLSServer() *TLSServer {
	return &TLSServer{}
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
