package interfaces

import (
	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertLoader allows the retrieval of certificates.
type CertLoader interface {
	// Load retrieves a certificate for a list of domains.
	// Returns a nil certificate if none is found.
	Load(domains []string) (*certs.Certificate, error)
}
