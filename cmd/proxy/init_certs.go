package main

import (
	"github.com/off-sync/platform-proxy/app/certs/cmd/gencert"
	"github.com/off-sync/platform-proxy/app/certs/qry/getcert"
	"github.com/off-sync/platform-proxy/infra/certgen"
	"github.com/off-sync/platform-proxy/infra/certstore"
	"github.com/off-sync/platform-proxy/infra/filesystem"
)

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
