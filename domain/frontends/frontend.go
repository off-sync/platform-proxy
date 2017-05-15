package frontends

import "time"

type Frontend struct {
	DomainName           string
	ServiceName          string
	Certificate          string
	PrivateKey           string
	CertificateExpiresAt time.Time
}
