package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"

	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/app/frontends/cmd/setcertificate"
	"github.com/off-sync/platform-proxy/common/logging"
	"github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/infra/acmestore"
	"github.com/off-sync/platform-proxy/infra/certgen"
	"github.com/off-sync/platform-proxy/infra/certstore"
	"github.com/off-sync/platform-proxy/infra/infraaws"
	infraTime "github.com/off-sync/platform-proxy/infra/time"
)

var (
	acmeStoreTable   string
	acmeAccountEmail string
	certStoreTable   string
	forceGeneration  bool
	frontendsTable   string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&acmeStoreTable, "acme-store-table", "a", "",
		"The name of the DynamoDB table in which the ACME accounts are stored.")

	RootCmd.PersistentFlags().StringVarP(&acmeAccountEmail, "acme-account-email", "e", "",
		"The email address used to lookup or register the ACME account.")

	RootCmd.PersistentFlags().StringVarP(&certStoreTable, "cert-store-table", "c", "",
		"The name of the DynamoDB table in which the certificates are stored.")

	RootCmd.PersistentFlags().BoolVar(&forceGeneration, "force", false,
		"The name of the DynamoDB table in which the frontends are stored.")

	RootCmd.PersistentFlags().StringVarP(&frontendsTable, "frontends-table", "f", "",
		"The name of the DynamoDB table in which the frontends are stored.")
}

// RootCmd defines the root command for the Off-Sync.com CertGen application.
var RootCmd = &cobra.Command{
	Use:              "certgen",
	Short:            "CertGen generates certificates for Off-Sync.com frontends",
	Long:             ``,
	PersistentPreRun: persistentPreRunRootCmd,
	Run:              runRootCmd,
}

var (
	getCertQry            *getcert.Qry
	genCertCmd            *gencert.Cmd
	setCertificateCommand setcertificate.Command
)

var log = logging.NewFromLogrus(logrus.New())

func persistentPreRunRootCmd(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	certStore, err := certstore.NewDynamoDBCertStore(sess, certStoreTable, infraTime.NewSystemTime())
	if err != nil {
		log.WithError(err).Fatal("creating new DynamodDB certificate store")
	}

	acmeStore, err := acmestore.NewDynamoDBACMEStore(sess, acmeStoreTable)
	if err != nil {
		log.WithError(err).Fatal("creating new DynamodDB ACME store")
	}

	acmeAccount, err := acmeStore.Load(certgen.LetsEncryptProductionEndpoint, acmeAccountEmail)
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

	setCertificateCommand, err = infraaws.NewDynamoDBSetCertificateCommand(sess, frontendsTable)
	if err != nil {
		panic(err)
	}
}

func runRootCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("invalid arguments: specify at least 1 domain name")
	}

	domainNames := args

	var cert *certs.Certificate
	var err error

	// do not check if force generation is true
	if !forceGeneration {
		log.
			WithField("domain_names", domainNames).
			Info("checking certificate store")

		cert, err = getCertQry.Execute(getcert.Model{Domains: domainNames})
		if err != nil {
			log.WithError(err).Fatal("checking certificate store")
		}
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

	log.
		WithField("frontend", domainNames[0]).
		WithField("expires_at", certificateExpiresAt).
		Info("setting frontend certificate")

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
