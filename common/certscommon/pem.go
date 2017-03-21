package certscommon

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const rsaPrivateKeyType = "RSA PRIVATE KEY"

// EncodeRSAPrivateKey encodes a RSA private key to PEM bytes.
func EncodeRSAPrivateKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  rsaPrivateKeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// DecodeRSAPrivateKey decodes a RSA private key from PEM bytes.
func DecodeRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	b, _ := pem.Decode(data)
	if b == nil {
		return nil, fmt.Errorf("PEM block not found")
	}

	if b.Type != rsaPrivateKeyType {
		return nil, fmt.Errorf("invalid PEM block type: %s", b.Type)
	}

	return x509.ParsePKCS1PrivateKey(b.Bytes)
}
