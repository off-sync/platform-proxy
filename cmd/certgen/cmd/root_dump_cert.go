package cmd

import (
	"crypto/x509"
	"time"

	certsCom "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/domain/certs"
)

func dumpCertificate(cert *certs.Certificate) time.Time {
	tlsCert, err := certsCom.ConvertToTLS(cert)
	if err != nil {
		log.WithError(err).Fatal("converting certificate")
	}

	var t time.Time

	for i, asn1Data := range tlsCert.Certificate {
		c, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			log.WithError(err).Error("parsing certificate")
		}

		if i == 0 {
			t = c.NotAfter
		}

		log.
			WithField("dns_names", c.DNSNames).
			WithField("common_name", c.Subject.CommonName).
			WithField("not_before", c.NotBefore).
			WithField("not_after", c.NotAfter).
			Info("certificate")
	}

	return t
}
