package certs

import "crypto/x509"

// CertStore allows storage and retrieval of certificates.
type CertStore interface {
	GetCert(domain string) (*x509.Certificate, error)
	StoreCert(domain string, cert *x509.Certificate) error
}
