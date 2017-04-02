package getconfig

import (
	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/domain/sites"
)

type Qry struct {
	provider interfaces.ConfigProvider
}

func New(provider interfaces.ConfigProvider) *Qry {
	return &Qry{
		provider: provider,
	}
}

func (q *Qry) Execute() ([]*sites.Backend, []*sites.Frontend, error) {
	backends, err := q.provider.GetBackends()
	if err != nil {
		return nil, nil, err
	}

	frontends, err := q.provider.GetFrontends()
	if err != nil {
		return nil, nil, err
	}

	return backends, frontends, nil
}
