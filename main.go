package main

import (
	"crypto/tls"
	"net/http"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/common/certscommon"
	"github.com/off-sync/platform-proxy/infra/certgen"
	"github.com/off-sync/platform-proxy/infra/certstore"
	"github.com/off-sync/platform-proxy/infra/filesystem"
)

var log = logrus.New()

var getCertQry *getcert.Qry
var genCertCmd *gencert.Cmd

func init() {
	// create infra implementations
	certFS, err := filesystem.NewLocalFileSystem(filesystem.Root("C:\\Temp\\LocalCertStore"))
	if err != nil {
		log.WithError(err).Fatal("creating certificates file system")
	}

	certStore := certstore.NewFileSystemCertStore(certFS)

	// certGen := certgen.NewSelfSigned()

	acmeFS, err := filesystem.NewLocalFileSystem(filesystem.Root("C:\\Temp\\AcmeFS"))
	if err != nil {
		log.WithError(err).Fatal("creating ACME file system")
	}

	certGen, err := certgen.NewAcme(acmeFS, certgen.LetsEncryptProductionEndpoint, "hosting@off-sync.com")
	if err != nil {
		panic(err)
	}

	// create certificate commands and queries
	getCertQry = getcert.New(certStore)
	genCertCmd = gencert.New(certGen, certStore)
}

func main() {
	getCertificateFunc := func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := chi.ServerName

		crt, err := getCertQry.Execute(getcert.Model{Domain: domain})
		if err != nil {
			return nil, err
		}

		if crt == nil {
			log.WithField("domain", domain).Info("generating certificate")

			crt, err = genCertCmd.Execute(gencert.Model{Domain: domain, KeyBits: 4096})
			if err != nil {
				return nil, err
			}
		}

		log.Info("loaded certificate")

		tlsCrt, err := certscommon.ConvertToTLS(crt)
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
		log.WithError(err).Fatal("listening and serving TLS")
	}
}
