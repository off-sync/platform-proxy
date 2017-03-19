package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

// SelfSignedCertGen implements CertGen and generates self-signed certificates.
type SelfSignedCertGen struct {
	RSABits int
}

// NewSelfSignedCertGen creates a new self-signed certificate generator.
func NewSelfSignedCertGen(rsaBits int) *SelfSignedCertGen {
	return &SelfSignedCertGen{
		RSABits: rsaBits,
	}
}

// GenCert creates a self-signed certificate.
func (g *SelfSignedCertGen) GenCert(domain string) (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, g.RSABits)

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
			CommonName:   domain,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{domain},

		IsCA: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}

	return tlsCert, nil
}
