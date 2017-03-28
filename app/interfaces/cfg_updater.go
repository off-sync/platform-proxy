package interfaces

import (
	"errors"

	"github.com/off-sync/platform-proxy/domain/sites"
)

// ErrUnknownDomain is returned when an action is performed on
// a unknown domain.
var ErrUnknownDomain = errors.New("unknown domain")

// ErrDuplicateDomain is returned when an action would result in
// a domain to be configured more than once.
var ErrDuplicateDomain = errors.New("duplicate domain")

// ConfigUpdater defines the interface through which the proxy
// configuration can be updated.
type ConfigUpdater interface {
	// RemoveAllSites removes all sites from the configuration.
	// It must always succeed.
	RemoveAllSites()

	// RemoveSite removes a site by providing one of its domains as a key.
	// Returns ErrUnknownDomain if no site was configured with this domain.
	RemoveSite(domain string) error

	// AddSite adds a site to the configuration.
	// Returns ErrDuplicateDomain if one of the site's domains is already
	// configured.
	AddSite(site *sites.Site) error
}
