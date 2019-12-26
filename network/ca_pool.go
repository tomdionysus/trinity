package network

import (
	"crypto/x509"
	"errors"
	"io/ioutil"

	"github.com/tomdionysus/trinity/util"
)

// CAPool struct represent a pool of x509 certificate
type CAPool struct {
	Pool   *x509.CertPool
	Logger *util.Logger
}

// NewCAPool Create and initialise a CAPool
func NewCAPool(logger *util.Logger) *CAPool {
	inst := &CAPool{
		Pool:   x509.NewCertPool(),
		Logger: logger,
	}

	return inst
}

// LoadPEM load a PEM file by name
func (cp *CAPool) LoadPEM(certFileName string) error {
	cabytes, err := ioutil.ReadFile(certFileName)
	if err != nil {
		cp.Logger.Error("CAPool", "Cannot Load CA File '%s': %s", certFileName, err.Error())
		return err
	}

	if !cp.Pool.AppendCertsFromPEM(cabytes) {
		cp.Logger.Error("CAPool", "Cannot Parse PEM CA File '%s'", certFileName)
		return errors.New("Cannot Parse CA File")
	}

	return nil
}
