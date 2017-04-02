package interfaces

import "github.com/off-sync/platform-proxy/domain/sites"

// ConfigProvider defines the interface through which configuration
// can be provided to the proxy.
type ConfigProvider interface {
	// GetNotificationChannel returns a channel through which
	// notifications can be pushed from the Configuration Provider.
	// If a 'true' value is received from this channel, the configuration
	// must be updated. 'False' values should be ignored, and can be
	// used to implement a heartbeat mechanism.
	GetNotificationChannel() chan<- bool

	// GetBackends returns all backends that should be configured.
	GetBackends() ([]*sites.Backend, error)

	// GetFrontends returns all frontends that should be configured.
	GetFrontends() ([]*sites.Frontend, error)
}
