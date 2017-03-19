package main

import (
	"crypto/tls"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/off-sync/platform-proxy/svc/certs"
)

var certGen = certs.NewSelfSignedCertGen(4096)

func main() {
	cert, err := certGen.GenCert("dev.bphorns.nl")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	srv := &http.Server{
		Addr:    ":8443",
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cert, nil
			},
		},
	}

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}
