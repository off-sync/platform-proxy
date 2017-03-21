package interfaces

import "github.com/off-sync/platform-proxy/domain/certs"

// CertGen allows the creation of new certificates based on a domain name.
type CertGen interface {
	GenCert(domain string, keyBits int) (*certs.Certificate, error)
}
