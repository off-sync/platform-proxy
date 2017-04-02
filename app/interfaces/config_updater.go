package interfaces

import (
	"errors"

	"github.com/off-sync/platform-proxy/domain/sites"
)

// ErrUnknownBackend is returned when an action is performed on
// an unknown backend.
var ErrUnknownBackend = errors.New("unknown backend")

// ErrDuplicateDomain is returned when an action would result in
// a domain to be configured more than once.
var ErrDuplicateDomain = errors.New("duplicate domain")

// ConfigUpdater defines the interface through which the proxy
// configuration can be updated.
type ConfigUpdater interface {
	// Update updates the configuration by replacing it with the provided
	// backends and frontends.
	// It returns ErrUnknownBackend if a frontend is included for which the backend is unknown.
	// It returns ErrDuplicateDomain if multiple frontends use the same domain.
	Update(backends []*sites.Backend, frontends []*sites.Frontend) error
}
