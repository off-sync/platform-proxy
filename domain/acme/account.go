package acme

import "github.com/xenolf/lego/acme"

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
