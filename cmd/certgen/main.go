package main

import (
	"os"

	"crypto/x509"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	certsCom "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/common/logging"
	"github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/infra/certgen"
	"github.com/off-sync/platform-proxy/infra/certstore"
	"github.com/off-sync/platform-proxy/infra/filesystem"
	"github.com/xenolf/lego/acme"
)

var log = logging.NewFromLogrus(logrus.New())

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

	// ACME logger
	acme.Logger = logging.NewStdLogAdapter(log)

	certGen, err := certgen.NewAcme(acmeFS, certgen.LetsEncryptProductionEndpoint, "hosting@off-sync.com")
	if err != nil {
		panic(err)
	}

	// create certificate commands and queries
	getCertQry = getcert.New(certStore)
	genCertCmd = gencert.New(certGen, certStore)
}

func main() {
	domains := os.Args[1:]
	if len(domains) < 1 {
		log.Fatal("missing domains: provide at least 1")
	}

	log.
		WithField("domains", domains).
		Info("checking certificate store")

	cert, err := getCertQry.Execute(getcert.Model{Domains: domains})
	if err != nil {
		log.WithError(err).Fatal("checking certificate store")
	}

	if cert != nil {
		log.Info("existing certificate found")

		dumpCertificate(cert)

		log.Info("saving to DynamoDB Certificate Store")

		sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
		if err != nil {
			log.WithError(err).Fatal("creating new session")
		}

		dynamoDBCertStore, err := certstore.NewDynamoDBCertStore(sess, "off-sync-qa-certificates")
		if err != nil {
			log.WithError(err).Fatal("creating new DynamoDB certificate store")
		}

		err = dynamoDBCertStore.Save(domains, cert)
		if err != nil {
			log.WithError(err).Fatal("saving to DynamoDB certificate store")
		}

		return
	}

	log.Info("generating certificate")

	cert, err = genCertCmd.Execute(gencert.Model{Domains: domains, KeyBits: 4096})
	if err != nil {
		log.WithError(err).Fatal("generating certificate")
	}

	dumpCertificate(cert)
}

func dumpCertificate(cert *certs.Certificate) {
	tlsCert, err := certsCom.ConvertToTLS(cert)
	if err != nil {
		log.WithError(err).Fatal("converting certificate")
	}

	for _, asn1Data := range tlsCert.Certificate {
		c, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			log.WithError(err).Error("parsing certificate")
		}

		log.
			WithField("dns_names", c.DNSNames).
			WithField("common_name", c.Subject.CommonName).
			WithField("not_before", c.NotBefore).
			WithField("not_after", c.NotAfter).
			Info("certificate")
	}
}
