package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"

	"sync"

	"github.com/off-sync/platform-proxy/common/certs"
	certsDom "github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/domain/frontends"
	"github.com/off-sync/platform-proxy/domain/services"
)

type proxy struct {
	sync.RWMutex
	router          *mux.Router
	serviceHandlers map[string]http.Handler
	certificates    map[string]*tls.Certificate
}

func newProxy() *proxy {
	return &proxy{}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.RLock()
	router := p.router
	p.RUnlock()

	router.ServeHTTP(w, r)
}

func (p *proxy) getCertificateFunc(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	p.RLock()
	defer p.RUnlock()

	if tlsCrt, found := p.certificates[chi.ServerName]; found {
		return tlsCrt, nil
	}

	return nil, nil
}

func (p *proxy) updateFrontends(frontends []*frontends.Frontend) error {
	router := mux.NewRouter()
	certificates := make(map[string]*tls.Certificate)

	for _, frontend := range frontends {
		// configure handler for this frontend
		handler, exists := p.serviceHandlers[frontend.ServiceName]
		if !exists {
			handler = http.NotFoundHandler()
		}

		router.Host(frontend.DomainName).Handler(handler)

		// add certificate for this frontend
		tlsCrt, err := certs.ConvertToTLS(&certsDom.Certificate{
			Certificate: []byte(frontend.Certificate),
			PrivateKey:  []byte(frontend.PrivateKey),
		})
		if err != nil {
			return err
		}

		certificates[frontend.DomainName] = tlsCrt
	}

	// swap in new router and certificates
	p.Lock()
	p.router = router
	p.certificates = certificates
	p.Unlock()

	return nil
}

func (p *proxy) updateServices(services []*services.Service) error {
	p.serviceHandlers = make(map[string]http.Handler)

	for _, service := range services {
		fwd, err := forward.New()
		if err != nil {
			return err
		}

		lb, err := roundrobin.New(fwd)
		if err != nil {
			return err
		}

		for _, server := range service.Servers {
			addrs, err := net.LookupHost(server.Hostname())
			if err != nil {
				return err
			}

			for _, addr := range addrs {
				u := &url.URL{}
				*u = *server
				u.Host = fmt.Sprintf("%s:%s", addr, server.Port())

				lb.UpsertServer(u)
			}
		}

		p.serviceHandlers[service.Name] = lb
	}

	return nil
}
