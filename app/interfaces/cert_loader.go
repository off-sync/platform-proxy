package interfaces

import (
	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertLoader allows the retrieval of certificates.
type CertLoader interface {
	// Load tries to retrieve a certificate for a list of domains.
	// It returns a nil certificate if it does not exist.
	Load(domains []string) (*certs.Certificate, error)
}
