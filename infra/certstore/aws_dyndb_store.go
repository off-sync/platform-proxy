package certstore

import (
	"fmt"
	"strings"
	"time"

	"crypto/sha256"

	"encoding/hex"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/off-sync/platform-proxy/app/interfaces"
	commonCerts "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/common/dyndbutil"
	"github.com/off-sync/platform-proxy/domain/certs"
	uuid "github.com/satori/go.uuid"
)

// DynamoDBCertStore is a certificate store implementation using Amazon DynamoDB.
// It is concurrent-safe, and provides auto-purging of expired certificates.
type DynamoDBCertStore struct {
	dyndbSvc  *dynamodb.DynamoDB
	tableName string
	time      interfaces.Time
}

// NewDynamoDBCertStore creates a new DynamoDB certificate store.
// It creates an AWS DynamoDB client and verifies whether the provided table exists.
func NewDynamoDBCertStore(p client.ConfigProvider, tableName string, time interfaces.Time) (*DynamoDBCertStore, error) {
	dyndbSvc := dynamodb.New(p)

	_, err := dyndbSvc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	return &DynamoDBCertStore{
		dyndbSvc:  dyndbSvc,
		tableName: tableName,
		time:      time,
	}, nil
}

type dynamoDBCertState int

const (
	dynamoDBCertStateNotLoaded = iota
	dynamoDBCertStateReserved
	dynamoDBCertStateGenerated
)

type dynamoDBCert struct {
	// Domains holds the domains for which this certificate is applicable.
	Domains []string

	// SaveToken is used to prevent race-conditions when a certificate needs to
	// be (re-)generated.
	SaveToken string

	// SaveTokenExpiresAt defines the time in UTC until which the current save token is valid.
	// After it expires each process can claim a new save token.
	SaveTokenExpiresAt time.Time

	// Created holds the time in UTC at which the certificate was added to the store.
	Created time.Time

	// Modified holds the time in UTC at which the certificate was last modified.
	Modified time.Time

	// PrivateKey contains the PEM encoded private key.
	PrivateKey string

	// Certificate contains the PEM encoded certificate.
	Certificate string

	// NotAfter holds the expiry date in UTC for the certificate.
	// This field is used for automatic cleanup by the DynamoDB TTL functionality.
	NotAfter time.Time
}

func (c *dynamoDBCert) hash() string {
	hash := sha256.Sum256([]byte(strings.Join(c.Domains, ",")))
	return hex.EncodeToString(hash[:])
}

func (s *DynamoDBCertStore) getItem(hash string, attrs ...string) (*dynamodb.GetItemOutput, error) {
	return s.dyndbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Hash": dyndbutil.StringAttr(hash),
		},
		AttributesToGet: aws.StringSlice(attrs),
	})
}

func (s *DynamoDBCertStore) putItem(item *dynamodb.PutItemInput) error {
	item.TableName = aws.String(s.tableName)
	_, err := s.dyndbSvc.PutItem(item)
	return err
}

func (s *DynamoDBCertStore) getCert(domains []string) (*dynamoDBCert, error) {
	c := &dynamoDBCert{
		Domains: domains,
	}

	i, err := s.getItem(c.hash(), "SaveToken", "SaveTokenExpiresAt", "Created", "Modified", "PrivateKey", "Certificate", "NotAfter")
	if err != nil {
		return nil, err
	}

	if i.Item == nil {
		// item not found
		return nil, nil
	}

	c.SaveToken = dyndbutil.StringValue(i.Item["SaveToken"])
	c.SaveTokenExpiresAt, err = dyndbutil.TimeValue(i.Item["SaveTokenExpiresAt"])
	if err != nil {
		return nil, err
	}

	c.Created, err = dyndbutil.TimeValue(i.Item["Created"])
	if err != nil {
		return nil, err
	}

	c.Modified, err = dyndbutil.TimeValue(i.Item["Modified"])
	if err != nil {
		return nil, err
	}

	c.PrivateKey = dyndbutil.StringValue(i.Item["PrivateKey"])
	c.Certificate = dyndbutil.StringValue(i.Item["Certificate"])
	c.NotAfter, err = dyndbutil.TimeValue(i.Item["NotAfter"])
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *DynamoDBCertStore) putCert(crt *dynamoDBCert) error {
	item := &dynamodb.PutItemInput{}

	now := s.time.Now()
	if crt.Created.IsZero() {
		// never saved before: set created
		crt.Created = now

		// check that this item does not exist
		item.ConditionExpression = aws.String("attribute_not_exists(#hash)")
		item.ExpressionAttributeNames = map[string]*string{
			"#hash": aws.String("Hash"),
		}
	} else {
		// already exists: set modified
		crt.Modified = now

		// check that the save token is expired, or match
		item.ConditionExpression = aws.String("(SaveTokenExpiresAt <= :now) or (SaveToken = :saveToken)")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":saveToken": dyndbutil.StringAttr(crt.SaveToken),
			":now":       dyndbutil.TimeAttr(now),
		}
	}

	item.Item = map[string]*dynamodb.AttributeValue{
		"Hash":               dyndbutil.StringAttr(crt.hash()),
		"Domains":            dyndbutil.StringListAttr(crt.Domains),
		"SaveToken":          dyndbutil.StringAttr(crt.SaveToken),
		"SaveTokenExpiresAt": dyndbutil.TimeAttr(crt.SaveTokenExpiresAt),
		"Created":            dyndbutil.TimeAttr(crt.Created),
		"Modified":           dyndbutil.TimeAttr(crt.Modified),
		"PrivateKey":         dyndbutil.StringAttr(crt.PrivateKey),
		"Certificate":        dyndbutil.StringAttr(crt.Certificate),
		"NotAfter":           dyndbutil.TimeAttr(crt.NotAfter),
	}

	return s.putItem(item)
}

// ClaimSaveToken tries to claim a save token in the store.
// ErrTokenAlreadyClaimed is returned if a non-expired token is already present.
func (s *DynamoDBCertStore) ClaimSaveToken(domains []string) (interfaces.CertSaveToken, error) {
	// check if there is an existing certificate in the store
	c, err := s.getCert(domains)
	if err != nil {
		return "", err
	}

	if c == nil {
		// certificate does not exist yet: create a new one
		c = &dynamoDBCert{
			Domains: domains,
		}
	}

	now := s.time.Now()

	if c.SaveToken != "" && c.SaveTokenExpiresAt.After(now) {
		// non-expired save token present: return ErrTokenAlreadyClaimed
		return "", interfaces.ErrTokenAlreadyClaimed
	}

	// create new save token and set expiry to 15 minutes from now
	c.SaveToken = uuid.NewV4().String()
	c.SaveTokenExpiresAt = now.Add(15 * time.Minute)

	if c.NotAfter.Before(c.SaveTokenExpiresAt) {
		// update if certificate is new or really close to expiring
		// to prevent cleaning by the DynamoDB TTL process
		c.NotAfter = c.SaveTokenExpiresAt
	}

	// save the certificate to the table
	err = s.putCert(c)
	if err != nil {
		return "", err
	}

	return interfaces.CertSaveToken(c.SaveToken), nil
}

// Save tries to save a certificate to the store. It is concurrent-safe.
func (s *DynamoDBCertStore) Save(domains []string, token interfaces.CertSaveToken, crt *certs.Certificate) error {
	// check if there is an existing certificate in the store
	c, err := s.getCert(domains)
	if err != nil {
		return err
	}

	if c == nil {
		// a certificate with a valid claim token should always already exist
		return fmt.Errorf("certificate for %v has not been claimed first", domains)
	}

	now := s.time.Now()
	if c.SaveToken != string(token) || c.SaveTokenExpiresAt.Before(now) {
		return interfaces.ErrInvalidSavetoken
	}

	if crt.PrivateKey != nil && crt.Certificate != nil {
		// copy the private key and certificate
		c.PrivateKey = string(crt.PrivateKey)
		c.Certificate = string(crt.Certificate)

		// get expiry date from certificate
		c.NotAfter, err = commonCerts.NotAfter(crt)
		if err != nil {
			return err
		}
	}

	err = s.putCert(c)
	if err != nil {
		if awsErr, isAws := err.(awserr.Error); isAws {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				// conditional check should only fail on an invalid save token
				return interfaces.ErrInvalidSavetoken
			}
		}

		return err
	}

	return nil
}

// Load tries to load an existing certificate from the store.
// If returns a nil certificate if it does not exists.
func (s *DynamoDBCertStore) Load(domains []string) (crt *certs.Certificate, err error) {
	c, err := s.getCert(domains)
	if err != nil {
		return nil, err
	}

	if c == nil || c.PrivateKey == "" || c.Certificate == "" {
		return nil, nil
	}

	return &certs.Certificate{
		PrivateKey:  []byte(c.PrivateKey),
		Certificate: []byte(c.Certificate),
	}, nil
}
