package sites

import (
	"net/url"
)

// Frontend maps a list of domain names to a Backend.
type Frontend struct {
	// Domain contains the domain name for this frontend.
	Domain string
	// BackendName specifies the name of the backend for this frontend.
	BackendName string
}

// NewFrontend creates a new frontend.
func NewFrontend(backendName string, domain string) *Frontend {
	return &Frontend{
		BackendName: backendName,
		Domain:      domain,
	}
}

// Backend defines a destination to which traffic can be served by the proxy.
type Backend struct {
	// Name holds the name of this site
	Name string
	// Servers contains a list of backend servers defined by the URL on
	// which they can be reached.
	Servers []*url.URL
}

// NewBackend creates a new backend. It tries to parse all provided servers to URLs.
func NewBackend(name string, servers ...string) (*Backend, error) {
	backend := &Backend{
		Name:    name,
		Servers: make([]*url.URL, len(servers)),
	}

	for i, server := range servers {
		u, err := url.Parse(server)
		if err != nil {
			return nil, err
		}

		backend.Servers[i] = u
	}

	return backend, nil
}
