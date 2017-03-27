package interfaces

import (
	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertSaver allows storage of certificates.
type CertSaver interface {
	// Save stores a certificate for a list of domains.
	Save(domains []string, crt *certs.Certificate) error
}
