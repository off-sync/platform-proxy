package certgen

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"

	"net/url"

	"github.com/off-sync/platform-proxy/common/certscommon"
	"github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/infra/filesystem"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/route53"
)

type AcmeUser struct {
	Email        string
	Registration *acme.RegistrationResource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

const (
	// LetsEncryptStagingEndpoint holds the URL of the Let's Encrypt staging endpoint.
	LetsEncryptStagingEndpoint = "https://acme-staging.api.letsencrypt.org/directory"

	// LetsEncryptProductionEndpoint holds the URL of the Let's Encrypt production endpoint.
	LetsEncryptProductionEndpoint = "https://acme-v01.api.letsencrypt.org/directory"
)

type AcmeCertGen struct {
	user   *AcmeUser
	client *acme.Client
}

const (
	regSuffix = "-reg.json"
	keySuffix = "-key.pem"
)

func getEmailPath(acmeEndpoint, email string) string {
	u, err := url.Parse(acmeEndpoint)
	if err != nil {
		panic(err)
	}

	acmeEndpoint = strings.Replace(u.Host, ".", "_", -1)

	email = strings.Replace(email, ".", "_", -1)

	return acmeEndpoint + "-" + email
}

// NewAcme creates a new ACME certificate generator. It creates and registers a new account
// if it not already exists.
func NewAcme(fs filesystem.FileSystem, acmeEndpoint string, email string) (*AcmeCertGen, error) {
	if email == "" {
		return nil, fmt.Errorf("empty email address")
	}

	if acmeEndpoint == "" {
		return nil, fmt.Errorf("empty ACME endpoint")
	}

	emailPath := getEmailPath(acmeEndpoint, email)

	keyPath := emailPath + keySuffix
	keyExists, err := fs.FileExists(keyPath)
	if err != nil {
		return nil, err
	}

	var key *rsa.PrivateKey

	if !keyExists {
		// generate private key
		key, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}

		keyBytes := certscommon.EncodeRSAPrivateKey(key)

		if err := fs.WriteBytes(keyPath, keyBytes); err != nil {
			return nil, err
		}
	} else {
		// load private key
		keyBytes, err := fs.ReadBytes(keyPath)
		if err != nil {
			return nil, err
		}

		key, err = certscommon.DecodeRSAPrivateKey(keyBytes)
		if err != nil {
			return nil, err
		}
	}

	regPath := emailPath + regSuffix
	regExists, err := fs.FileExists(regPath)
	if err != nil {
		return nil, err
	}

	acmeUser := &AcmeUser{
		Email: email,
		key:   key,
	}

	acmeClient, err := acme.NewClient(acmeEndpoint, acmeUser, acme.RSA4096)
	if err != nil {
		return nil, err
	}

	if keyExists && regExists {
		// load existing registration
		regBytes, err := fs.ReadBytes(regPath)
		if err != nil {
			return nil, err
		}

		acmeUser.Registration = &acme.RegistrationResource{}
		json.Unmarshal(regBytes, acmeUser.Registration)

		// check email address
		contact := acmeUser.Registration.Body.Contact
		if len(contact) != 1 || contact[0] != "mailto:"+email {
			return nil, fmt.Errorf(
				"account file for '%s' contains different email address: %s",
				email, contact[0])
		}
	} else {
		// register new account
		acmeUser.Registration, err = acmeClient.Register()
		if err != nil {
			return nil, err
		}

		// agree to TOS
		err = acmeClient.AgreeToTOS()
		if err != nil {
			return nil, err
		}

		// save reg
		regBytes, err := json.Marshal(acmeUser.Registration)
		if err != nil {
			return nil, err
		}

		if err := fs.WriteBytes(regPath, regBytes); err != nil {
			return nil, err
		}
	}

	provider, err := route53.NewDNSProvider()
	if err != nil {
		return nil, err
	}

	acmeClient.SetChallengeProvider(acme.DNS01, provider)
	acmeClient.ExcludeChallenges([]acme.Challenge{acme.TLSSNI01, acme.HTTP01})

	return &AcmeCertGen{
		user:   acmeUser,
		client: acmeClient,
	}, nil
}

// GenCert generates a certificate using the provided ACME endpoint.
func (g *AcmeCertGen) GenCert(domain string, keyBits int) (*certs.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, fmt.Errorf("generating RSA private key: %s", err)
	}

	crt, failures := g.client.ObtainCertificate([]string{domain}, true, key, false)
	if len(failures) > 0 {
		return nil, fmt.Errorf("obtaining certificates: %v", failures)
	}

	return &certs.Certificate{
		Certificate: crt.Certificate,
		PrivateKey:  crt.PrivateKey,
	}, nil
}
