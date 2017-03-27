package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	certsDom "github.com/off-sync/platform-proxy/domain/certs"
)

// ConvertToTLS converts an internal certificate to the format
// used by the crypto/tls package.
func ConvertToTLS(crt *certsDom.Certificate) (*tls.Certificate, error) {
	var certs [][]byte

	var p *pem.Block
	for rest := crt.Certificate; len(rest) > 0; {
		if p, rest = pem.Decode(rest); p != nil {
			certs = append(certs, p.Bytes)
		}
	}

	if len(certs) < 1 {
		return nil, fmt.Errorf("unable to decode certificate(s)")
	}

	p, _ = pem.Decode(crt.PrivateKey)
	if p == nil {
		return nil, fmt.Errorf("unable to decode private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %s", err)
	}

	return &tls.Certificate{
		Certificate: certs,
		PrivateKey:  key,
	}, nil
}
