package interfaces

import (
	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertLoader allows the retrieval of certificates.
type CertLoader interface {
	// LoadOrGenerate tries to retrieve a certificate for a list of domains.
	// If it does not exist yet, the provide certificate generator is used to
	// create a new certificate.
	LoadOrGenerate(domains []string, gen CertGen) (*certs.Certificate, error)
}
