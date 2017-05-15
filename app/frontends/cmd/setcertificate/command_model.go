package setcertificate

import "time"

type CommandModel struct {
	DomainName           string
	Certificate          string
	PrivateKey           string
	CertificateExpiresAt time.Time
}
