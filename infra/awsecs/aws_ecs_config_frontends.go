package awsecs

import (
	"fmt"

	"github.com/off-sync/platform-proxy/domain/sites"
)

// GetFrontends returns a list of frontends based on the backends.
func (p *ConfigProvider) GetFrontends() ([]*sites.Frontend, error) {
	var frontends []*sites.Frontend

	backends, err := p.GetBackends()
	if err != nil {
		return nil, err
	}

	for _, backend := range backends {
		frontends = append(frontends, sites.NewFrontend(
			backend.Name,
			fmt.Sprintf("%s.qa.off-sync.net", backend.Name)))
	}

	return frontends, nil
}
