package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/common/logging"
	"github.com/off-sync/platform-proxy/domain/acme"
	"github.com/off-sync/platform-proxy/domain/certs"
	lego "github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/route53"
)

const (
	// LetsEncryptStagingEndpoint holds the URL of the Let's Encrypt staging endpoint.
	LetsEncryptStagingEndpoint = "https://acme-staging.api.letsencrypt.org/directory"

	// LetsEncryptProductionEndpoint holds the URL of the Let's Encrypt production endpoint.
	LetsEncryptProductionEndpoint = "https://acme-v01.api.letsencrypt.org/directory"
)

// LegoACMECertGen implements the CertGen interface using the
// lego ACME library.
type LegoACMECertGen struct {
	user    *acme.Account
	client  *lego.Client
	keyBits int
}

// SetLegoLogger sets the logger used by the lego library.
func SetLegoLogger(log interfaces.Logger) error {
	if log == nil {
		return fmt.Errorf("missing logger")
	}

	lego.Logger = logging.NewStdLogAdapter(log)

	return nil
}

// NewLegoACMECertGen creates a new ACME certificate generator for the provided account.
func NewLegoACMECertGen(account *acme.Account, log interfaces.Logger) (*LegoACMECertGen, error) {
	if account == nil {
		return nil, fmt.Errorf("missing account")
	}

	if log != nil {
		SetLegoLogger(log)
	}

	acmeClient, err := lego.NewClient(account.Endpoint, account, lego.RSA4096)
	if err != nil {
		return nil, err
	}

	provider, err := route53.NewDNSProvider()
	if err != nil {
		return nil, err
	}

	acmeClient.SetChallengeProvider(lego.DNS01, provider)
	acmeClient.ExcludeChallenges([]lego.Challenge{lego.TLSSNI01, lego.HTTP01})

	return &LegoACMECertGen{
		user:    account,
		client:  acmeClient,
		keyBits: 4096,
	}, nil
}

// GenCert generates a certificate using the provided ACME endpoint.
func (g *LegoACMECertGen) GenCert(domains []string) (*certs.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, g.keyBits)
	if err != nil {
		return nil, fmt.Errorf("generating RSA private key: %s", err)
	}

	crt, failures := g.client.ObtainCertificate(domains, true, key, false)
	if len(failures) > 0 {
		return nil, fmt.Errorf("obtaining certificates: %v", failures)
	}

	return &certs.Certificate{
		Certificate: crt.Certificate,
		PrivateKey:  crt.PrivateKey,
	}, nil
}
