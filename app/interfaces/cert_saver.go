package interfaces

import (
	"errors"

	"github.com/off-sync/platform-proxy/domain/certs"
)

// CertSaveToken is an opaque token used to save certificates.
type CertSaveToken string

// ErrTokenAlreadyClaimed is returned by GetSaveToken when there
// already is an active save token claimed for the provided domains.
var ErrTokenAlreadyClaimed = errors.New("token already claimed")

// ErrInvalidSavetoken is returned by Save when the provided token
// is invalid (non-existing, expired).
var ErrInvalidSavetoken = errors.New("invalid save token")

// CertSaver allows storage of certificates.
type CertSaver interface {
	// ClaimSaveToken tries to claim a save token for the provided domains.
	// This can be used to implement a concurrent-safe implementation.
	// It returns ErrTokenAlreadyClaimed if the token could not be claimed.
	ClaimSaveToken(domains []string) (CertSaveToken, error)

	// Save stores a certificate for a list of domains.
	// Requires an active save token.
	// It returns ErrInvalidSavetoken if an invalid save token is provided.
	Save(domains []string, token CertSaveToken, crt *certs.Certificate) error
}
