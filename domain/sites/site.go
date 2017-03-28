package sites

import (
	"net/url"
)

// Site defines a site for which traffic is served by the proxy.
type Site struct {
	// Domains includes the applicable domains for this site.
	Domains []string
	// Backends contains a list of backend servers defined by the URL on
	// which they can be reached.
	Backends []*url.URL
}

// New creates a new site. It tries to parse all provided backends to URLs.
func New(domains []string, backends []string) (*Site, error) {
	site := &Site{
		Domains:  make([]string, len(domains)),
		Backends: make([]*url.URL, len(backends)),
	}

	copy(site.Domains, domains)

	for i, backend := range backends {
		u, err := url.Parse(backend)
		if err != nil {
			return nil, err
		}

		site.Backends[i] = u
	}

	return site, nil
}
