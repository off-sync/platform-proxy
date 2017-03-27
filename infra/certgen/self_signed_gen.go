package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"github.com/off-sync/platform-proxy/domain/certs"
)

// SelfSignedCertGen implements CertGen and generates self-signed certificates.
type SelfSignedCertGen struct {
}

// NewSelfSigned creates a new self-signed certificate generator.
func NewSelfSigned() *SelfSignedCertGen {
	return &SelfSignedCertGen{}
}

// GenCert creates a self-signed certificate.
func (g *SelfSignedCertGen) GenCert(domains []string, keyBits int) (*certs.Certificate, error) {
	if len(domains) < 1 {
		return nil, fmt.Errorf("domains missing: provide at least 1 domain")
	}

	priv, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now().UTC()

	notAfter := time.Now().UTC().Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Off-Sync.com"},
			CommonName:   domains[0],
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: domains,

		IsCA: true,
	}

	crtBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(priv)

	return &certs.Certificate{
		Certificate: crtBytes,
		PrivateKey:  keyBytes,
	}, nil
}
