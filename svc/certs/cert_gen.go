package certs

import "crypto/tls"

// CertGen allows the creation of new certificates based on a domain name.
type CertGen interface {
	GenCert(domain string) (*tls.Certificate, error)
}
