package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/common/logging"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

var log = logging.NewFromLogrus(logrus.New())

func main() {
	getCertificateFunc := func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domains := make([]string, 1)
		domains[0] = chi.ServerName

		crt, err := getCertQry.Execute(getcert.Model{Domains: domains})
		if err != nil {
			return nil, err
		}

		if crt == nil {
			log.
				WithField("domains", domains).
				Info("generating certificate")

			crt, err = genCertCmd.Execute(gencert.Model{Domains: domains})
			if err != nil {
				return nil, err
			}
		}

		log.
			WithField("domains", domains).
			Info("loaded certificate")

		tlsCrt, err := certs.ConvertToTLS(crt)
		if err != nil {
			return nil, err
		}

		return tlsCrt, nil
	}

	backends, frontends, err := getConfigQry.Execute()
	if err != nil {
		log.WithError(err).Fatal("getting configuration")
	}

	r := mux.NewRouter()

	backendHandlers := make(map[string]http.Handler)
	for _, backend := range backends {
		fwd, err := forward.New()
		if err != nil {
			log.WithError(err).Fatal("creating forwarder")
		}

		lb, err := roundrobin.New(fwd)
		if err != nil {
			log.WithError(err).Fatal("creating load balancer")
		}

		for _, server := range backend.Servers {
			addrs, err := net.LookupHost(server.Hostname())
			if err != nil {
				log.
					WithField("server", server).
					WithError(err).
					Fatal("looking up server host")
			}

			for _, addr := range addrs {
				u := &url.URL{}
				*u = *server
				u.Host = fmt.Sprintf("%s:%s", addr, server.Port())

				log.
					WithField("server", server).
					WithField("addr", u).
					Info("adding server")

				lb.UpsertServer(u)
			}
		}

		backendHandlers[backend.Name] = lb
	}

	for _, frontend := range frontends {
		handler, exists := backendHandlers[frontend.BackendName]
		if !exists {
			log.WithField("backend_name", frontend.BackendName).Fatal("unknown backend name")
		}

		r.Host(frontend.Domain).Handler(handler)
	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>%s</h1>\n<pre>", r.Host)

		for name, values := range r.Header {
			fmt.Fprintf(w, "%s: %v\n", name, values)
		}

		fmt.Fprint(w, "</pre>\n")
	})

	srv := &http.Server{
		Addr:    ":8443",
		Handler: r,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			GetCertificate: getCertificateFunc,
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.
			WithError(err).
			Fatal("listening and serving TLS")
	}
}
