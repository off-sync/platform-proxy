package updatecfg

import (
	"github.com/off-sync/platform-proxy/domain/sites"
)

// Model defines the input for the Update Config command.
type Model struct {
	Backends  []*sites.Backend
	Frontends []*sites.Frontend
}
