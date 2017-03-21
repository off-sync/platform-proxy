package certs

// Certificate defines a certificate. Both Certificate and PrivateKey hold
// byte arrays of PEM encoded data.
type Certificate struct {
	Certificate []byte
	PrivateKey  []byte
}
