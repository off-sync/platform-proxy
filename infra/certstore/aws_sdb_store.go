package certstore

import (
	"fmt"
	"strconv"
	"strings"

	"time"

	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/off-sync/platform-proxy/domain/certs"
)

// SimpleDBCertStore implements a certificate loader and saver using
// AWS SimpleDB as its backend. It is concurrent.
type SimpleDBCertStore struct {
	sdbSvc     *simpledb.SimpleDB
	domainName string
}

// NewSimpleDBCertStore returns a new SimpleDBCertStore. It uses the provided AWS config provider
// to create a SimpleDB client. The provided domain name is checked by performing a call to
// DomainMetadata.
func NewSimpleDBCertStore(p client.ConfigProvider, domainName string) (*SimpleDBCertStore, error) {
	sdbSvc := simpledb.New(p)

	// query domain data to check whether this is a valid domain name
	_, err := sdbSvc.DomainMetadata(&simpledb.DomainMetadataInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		return nil, err
	}

	return &SimpleDBCertStore{
		sdbSvc:     sdbSvc,
		domainName: domainName,
	}, nil
}

type simpleDBCertState int

const (
	simpleDBCertStateNotLoaded = iota
	simpleDBCertStateReserved
	simpleDBCertStateGenerated
)

type simpleDBCert struct {
	// Domains is used to determine the unique item name.
	Domains []string
	// Version is used for optimistic locking.
	Version int
	// State determines the state of the certificate.
	State       simpleDBCertState
	Created     time.Time
	Modified    time.Time
	PrivateKey  string
	Certificate string
}

func newSimpleDBCert(domains []string) *simpleDBCert {
	return &simpleDBCert{
		Domains: domains,
		State:   simpleDBCertStateNotLoaded,
	}
}

func (c *simpleDBCert) itemName() string {
	return strings.Join(c.Domains, ",")
}

func (c *simpleDBCert) certificate() *certs.Certificate {
	if c.PrivateKey == "" || c.Certificate == "" {
		return nil
	}

	return &certs.Certificate{
		PrivateKey:  []byte(c.PrivateKey),
		Certificate: []byte(c.Certificate),
	}
}

func (c *simpleDBCert) load(s *SimpleDBCertStore) error {
	if c.State != simpleDBCertStateNotLoaded {
		return fmt.Errorf("state not allowed when loading: %v", c.State)
	}

	sel, err := s.sdbSvc.Select(&simpledb.SelectInput{
		SelectExpression: aws.String(fmt.Sprintf(
			"select * from `%s` where itemName() = `%s`",
			s.domainName, c.itemName())),
	})
	if err != nil {
		return err
	}

	if len(sel.Items) < 1 {
		// not found
		return nil
	}

	if len(sel.Items) > 1 {
		return fmt.Errorf("found duplicate certificates for item name: %s", c.itemName())
	}

	for _, a := range sel.Items[0].Attributes {
		if *a.Name == "Certificate" {
			// ignore all other attributes
			continue
		}

		return json.Unmarshal([]byte(*a.Value), c)
	}

	return fmt.Errorf("Certificate attribute not found in select: %v", sel)
}

func (c *simpleDBCert) save(s *SimpleDBCertStore, cert *certs.Certificate) error {
	if cert == nil {
		// if no certificate is provided, this is stored as a reserved state.
		// this prevents race conditions when generating new certificates which can
		// take some time to complete.
		c.State = simpleDBCertStateReserved
		c.PrivateKey = ""
		c.Certificate = ""
	} else {
		c.State = simpleDBCertStateGenerated
		c.PrivateKey = string(cert.PrivateKey)
		c.Certificate = string(cert.Certificate)
	}

	exp := &simpledb.UpdateCondition{Name: aws.String("Version")}
	if c.Version == 0 {
		// never saved before: Version must not exist
		exp.Exists = aws.Bool(false)

		c.Created = time.Now().UTC()
	} else {
		// optimistic locking: version must be equal to previous version
		exp.Value = aws.String(strconv.Itoa(c.Version - 1))

		c.Modified = time.Now().UTC()
	}

	// increment version so that this is included when marshalling to JSON
	c.Version++

	attrs := make([]*simpledb.ReplaceableAttribute, 2)

	// Version
	attrs[0] = &simpledb.ReplaceableAttribute{
		Name:    aws.String("Version"),
		Value:   aws.String(strconv.Itoa(c.Version)),
		Replace: aws.Bool(true),
	}

	// Data
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	attrs[1] = &simpledb.ReplaceableAttribute{
		Name:    aws.String("Data"),
		Value:   aws.String(string(data)),
		Replace: aws.Bool(true),
	}

	_, err = s.sdbSvc.PutAttributes(&simpledb.PutAttributesInput{
		DomainName: aws.String(s.domainName),
		ItemName:   aws.String(c.itemName()),
		Attributes: attrs,
		Expected:   exp,
	})

	return err
}

// Load tries to load a certificate from the store. It returns a nil certificate if not found.
//  It is concurrent.
func (s *SimpleDBCertStore) Load(domains []string) (*certs.Certificate, error) {
	simpleDBCert := newSimpleDBCert(domains)

	err := simpleDBCert.load(s)
	if err != nil {
		return nil, err
	}

	return simpleDBCert.certificate(), nil
}

// Save tries to save a certificate to the store. It is concurrent.
func (s *SimpleDBCertStore) Save(domains []string, crt *certs.Certificate) error {
	simpleDBCert := newSimpleDBCert(domains)

	return simpleDBCert.save(s, crt)
}
