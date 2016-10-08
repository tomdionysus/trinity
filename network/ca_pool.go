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

func (me *CAPool) LoadPEM(certFile string) error {
	cabytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		me.Logger.Error("CAPool", "Cannot Load CA File '%s': %s", certFile, err.Error())
		return err
	}

	if !me.Pool.AppendCertsFromPEM(cabytes) {
		me.Logger.Error("CAPool", "Cannot Parse PEM CA File '%s'", certFile)
		return errors.New("Cannot Parse CA File")
	}

	return nil
}
