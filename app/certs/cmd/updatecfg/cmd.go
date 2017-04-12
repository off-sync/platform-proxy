package updatecfg

import "github.com/off-sync/platform-proxy/app/interfaces"

// Cmd defines the Update Config command.
type Cmd struct {
	cfgUpdater interfaces.ConfigUpdater
}

// New creates a new Update Config command using the provided
// Config Updater.
func New(cfgUpdater interfaces.ConfigUpdater) *Cmd {
	return &Cmd{
		cfgUpdater: cfgUpdater,
	}
}

// Execute executes the Update Config command.
func (c *Cmd) Execute(model *Model) error {
	return c.cfgUpdater.Update(model.Backends, model.Frontends)
}
