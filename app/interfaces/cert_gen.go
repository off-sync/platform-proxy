package interfaces

import "github.com/off-sync/platform-proxy/domain/certs"

// CertGen allows the creation of new certificates based on a list of domains.
type CertGen interface {
	GenCert(domains []string) (*certs.Certificate, error)
}
