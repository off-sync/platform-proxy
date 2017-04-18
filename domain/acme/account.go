package acme

import (
	"crypto"

	certsCom "github.com/off-sync/platform-proxy/common/certs"
	"github.com/xenolf/lego/acme"
)

// Account defines the required fields of an ACME account.
type Account struct {
	// Endpoint holds the ACME endpoint against which this account has been registered.
	Endpoint string

	// Email holds the email address used to register this account.
	Email string

	// PrivateKey contains the PEM encoded private key for this account.
	PrivateKey string

	// Registration contains the ACME registration resource.
	Registration *acme.RegistrationResource
}

// GetEmail returns the email address of this account.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetRegistration returns the registration resource of this account.
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

// GetPrivateKey returns the private key of this account as a
// crypto.PrivateKey. It panics if the key cannot be decoded.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	key, err := certsCom.DecodeRSAPrivateKey([]byte(a.PrivateKey))
	if err != nil {
		panic(err)
	}

	return key
}
