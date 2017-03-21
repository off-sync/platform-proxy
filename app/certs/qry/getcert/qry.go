package getcert

import (
	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/domain/certs"
)

// Qry defines the Get Certificate query.
type Qry struct {
	ldr interfaces.CertLoader
}

// New creates a new Get Certificate query using the provided certificate loader.
func New(ldr interfaces.CertLoader) *Qry {
	return &Qry{
		ldr: ldr,
	}
}

// Model defines the input for the Get Certificate query.
type Model struct {
	Domain string
}

// Execute performs the lookup of a certificate. Returns a nil certificate if not found.
func (q *Qry) Execute(model Model) (*certs.Certificate, error) {
	return q.ldr.Load(model.Domain)
}
