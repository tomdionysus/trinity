package network

import (
	"crypto/x509"
	"errors"
	"github.com/tomdionysus/trinity/util"
	"io/ioutil"
)

type CAPool struct {
	Pool   *x509.CertPool
	Logger *util.Logger
}

func NewCAPool(logger *util.Logger) *CAPool {
	inst := &CAPool{
		Pool:   x509.NewCertPool(),
		Logger: logger,
	}

	return inst
}

func (cp *CAPool) LoadPEM(certFile string) error {
	cabytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		cp.Logger.Error("CAPool", "Cannot Load CA File '%s': %s", certFile, err.Error())
		return err
	}

	if !cp.Pool.AppendCertsFromPEM(cabytes) {
		cp.Logger.Error("CAPool", "Cannot Parse PEM CA File '%s'", certFile)
		return errors.New("Cannot Parse CA File")
	}

	return nil
}
