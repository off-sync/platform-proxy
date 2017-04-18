package interfaces

import "github.com/off-sync/platform-proxy/domain/acme"

// ACMERegistrar provides the functionality to register new accounts
// with an ACME provider.
type ACMERegistrar interface {
	Register(endpoint, email string) (*acme.Account, error)
}

// ACMESaver provides a method to save an ACME account.
type ACMESaver interface {
	Save(account *acme.Account) error
}

// ACMELoader provides a method to load an ACME account based on the
// provided endpoint and email address.
// It returns a nil account if it does not exist.
type ACMELoader interface {
	Load(endpoint, email string) (*acme.Account, error)
}
