package main

import (
	"crypto/tls"
	"net/http"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/common/certs"
)

var log = logrus.New()

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

			crt, err = genCertCmd.Execute(gencert.Model{Domains: domains, KeyBits: 4096})
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

	r := mux.NewRouter()
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
