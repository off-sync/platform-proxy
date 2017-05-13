package services

import "net/url"

type Service struct {
	Name    string
	Servers []*url.URL
}

// NewService creates a new service. It tries to parse all provided servers to URLs.
func NewService(name string, servers ...string) (*Service, error) {
	service := &Service{
		Name:    name,
		Servers: make([]*url.URL, len(servers)),
	}

	for i, server := range servers {
		u, err := url.Parse(server)
		if err != nil {
			return nil, err
		}

		service.Servers[i] = u
	}

	return service, nil
}
