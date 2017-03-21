package interfaces

import (
	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertSaver allows storage of certificates.
type CertSaver interface {
	// Save stores a certificate for a domain.
	Save(domain string, crt *certs.Certificate) error
}
