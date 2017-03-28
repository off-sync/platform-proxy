package updatecfg

import (
	"github.com/off-sync/platform-proxy/domain/sites"
)

// Model defines the input for the Update Config command.
type Model struct {
	Sites []*sites.Site
}
