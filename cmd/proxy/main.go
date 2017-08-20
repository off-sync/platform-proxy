package main

import (
	"crypto/tls"
	"net/http"

	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/common/logging"
)

var log interfaces.Logger

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logr := logrus.New()
	logr.Formatter = &logrus.JSONFormatter{}

	log = logging.NewFromLogrus(logr)
}

func main() {
	proxy := newProxy()

	services, err := getServicesQuery.Execute(nil)
	if err != nil {
		log.WithError(err).Fatal("getting services")
	}

	err = proxy.updateServices(services.Services)
	if err != nil {
		log.WithError(err).Fatal("updating services")
	}

	frontends, err := getFrontendsQuery.Execute(nil)
	if err != nil {
		log.WithError(err).Fatal("getting frontends")
	}

	proxy.updateFrontends(frontends.Frontends)
	if err != nil {
		log.WithError(err).Fatal("updating frontends")
	}

	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := &url.URL{}
			*url = *r.URL
			url.Scheme = "https"
			url.Host = r.Host

			log.
				WithField("request_url", r.URL).
				WithField("redirect_url", url).
				Info("redirecting")

			w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		}),
	}

	go func() {
		log.WithField("addr", srv.Addr).Info("starting HTTP server")

		if err := srv.ListenAndServe(); err != nil {
			log.
				WithError(err).
				Fatal("listening and serving")
		}
	}()

	tlsSrv := &http.Server{
		Addr:    ":8443",
		Handler: proxy,
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
			GetCertificate: proxy.getCertificateFunc,
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	log.WithField("addr", tlsSrv.Addr).Info("starting HTTPS server")

	if err := tlsSrv.ListenAndServeTLS("", ""); err != nil {
		log.
			WithError(err).
			Fatal("listening and serving TLS")
	}
}
