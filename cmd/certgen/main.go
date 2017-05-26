package main

import (
	"os"
	"time"

	"crypto/x509"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/app/frontends/cmd/setcertificate"
	certsCom "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/common/logging"
	"github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/infra/acmestore"
	"github.com/off-sync/platform-proxy/infra/certgen"
	"github.com/off-sync/platform-proxy/infra/certstore"
	"github.com/off-sync/platform-proxy/infra/infraaws"
	infraTime "github.com/off-sync/platform-proxy/infra/time"
)

var log = logging.NewFromLogrus(logrus.New())

var getCertQry *getcert.Qry
var genCertCmd *gencert.Cmd
var setCertificateCommand setcertificate.Command

func init() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	certStore, err := certstore.NewDynamoDBCertStore(sess, "off-sync-qa-certificates", infraTime.NewSystemTime())
	if err != nil {
		log.WithError(err).Fatal("creating new DynamodDB certificate store")
	}

	acmeStore, err := acmestore.NewDynamoDBACMEStore(sess, "off-sync-qa-acme-account")
	if err != nil {
		log.WithError(err).Fatal("creating new DynamodDB ACME store")
	}

	acmeAccount, err := acmeStore.Load(certgen.LetsEncryptProductionEndpoint, "hosting@off-sync.com")
	if err != nil {
		log.WithError(err).Fatal("loading ACME account")
	}

	certGen, err := certgen.NewLegoACMECertGen(acmeAccount, log)
	if err != nil {
		panic(err)
	}

	// create certificate commands and queries
	getCertQry = getcert.New(certStore)
	genCertCmd = gencert.New(certGen, certStore)

	setCertificateCommand, err = infraaws.NewDynamoDBSetCertificateCommand(sess, "off-sync-qa-frontends")
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("invalid arguments: specify at least 1 domain name")
	}

	domainNames := os.Args[1:]

	log.
		WithField("domain_names", domainNames).
		Info("checking certificate store")

	cert, err := getCertQry.Execute(getcert.Model{Domains: domainNames})
	if err != nil {
		log.WithError(err).Fatal("checking certificate store")
	}

	if cert != nil {
		log.Info("existing certificate found")
	} else {
		log.Info("generating certificate")

		cert, err = genCertCmd.Execute(gencert.Model{Domains: domainNames})
		if err != nil {
			log.WithError(err).Fatal("generating certificate")
		}
	}

	certificateExpiresAt := dumpCertificate(cert)

	err = setCertificateCommand.Execute(&setcertificate.CommandModel{
		DomainName:           domainNames[0],
		Certificate:          string(cert.Certificate),
		PrivateKey:           string(cert.PrivateKey),
		CertificateExpiresAt: certificateExpiresAt,
	})
	if err != nil {
		log.WithError(err).Fatal("setting frontend certificate")
	}
}

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
