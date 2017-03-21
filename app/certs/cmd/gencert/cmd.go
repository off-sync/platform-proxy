package gencert

import (
	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/domain/certs"
)

// Cmd defines the Generate Certificate command.
type Cmd struct {
	gen interfaces.CertGen
	svr interfaces.CertSaver
}

// New creates a new Generate Certificate command with the provided options.
func New(gen interfaces.CertGen, svr interfaces.CertSaver) *Cmd {
	return &Cmd{
		gen: gen,
		svr: svr,
	}
}

// Model defines the input for the Generate Certificate command.
type Model struct {
	Domain  string
	KeyBits int
}

// Execute executes the Generate Certificate command.
// It generates a new certificate for the domain and stores it.
func (c *Cmd) Execute(model Model) (*certs.Certificate, error) {
	crt, err := c.gen.GenCert(model.Domain, model.KeyBits)
	if err != nil {
		return nil, err
	}

	err = c.svr.Save(model.Domain, crt)
	if err != nil {
		return nil, err
	}

	return crt, nil
}
