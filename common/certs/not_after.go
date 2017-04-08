package certs

import (
	"crypto/x509"
	"fmt"
	"time"

	certsDom "github.com/off-sync/platform-proxy/domain/certs"
)

var emptyTime = time.Time{}

// NotAfter tries to parse the certificate and returns the NotAfter property
// if successful.
func NotAfter(crt *certsDom.Certificate) (time.Time, error) {
	tlsCrt, err := ConvertToTLS(crt)
	if err != nil {
		return emptyTime, err
	}

	if len(tlsCrt.Certificate) < 1 {
		return emptyTime, fmt.Errorf("no certificates found: %s", string(crt.Certificate))
	}

	asnCrt, err := x509.ParseCertificate(tlsCrt.Certificate[0])
	if err != nil {
		return emptyTime, err
	}

	return asnCrt.NotAfter, nil
}
