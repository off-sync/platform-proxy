package certstore

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"encoding/json"

	"encoding/hex"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	commonCerts "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/domain/certs"
)

var (
	errUnableToReserveCertificate = errors.New("unable to reserve certificate")
	errUnableToUpdateCertificate  = errors.New("unable to update certificate")
)

type DynamoDBCertStore struct {
	dyndbSvc  *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBCertStore(p client.ConfigProvider, tableName string) (*DynamoDBCertStore, error) {
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
	}, nil
}

type dynamoDBCertState int

const (
	dynamoDBCertStateNotLoaded = iota
	dynamoDBCertStateReserved
	dynamoDBCertStateGenerated
)

type dynamoDBCert struct {
	// Domains is used to determine the unique item name.
	Domains []string
	// Version is used for optimistic locking.
	Version int
	// State determines the state of the certificate.
	State       dynamoDBCertState
	Created     time.Time
	Modified    time.Time
	PrivateKey  string
	Certificate string
	// NotAfter holds the expiry date for the certificate
	NotAfter time.Time
}

func hash(domains []string) string {
	hash := sha256.Sum256([]byte(strings.Join(domains, ",")))
	return hex.EncodeToString(hash[:])
}

func (s *DynamoDBCertStore) getItem(domains []string) (*dynamodb.GetItemOutput, error) {
	return s.dyndbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Hash": &dynamodb.AttributeValue{S: aws.String(hash(domains))},
		},
		AttributesToGet: []*string{
			aws.String("Data"),
		},
	})
}

func (s *DynamoDBCertStore) getCert(domains []string) (*dynamoDBCert, error) {
	i, err := s.getItem(domains)
	if err != nil {
		return nil, err
	}

	if i.Item == nil {
		// item not found
		return nil, nil
	}

	c := &dynamoDBCert{}

	data, found := i.Item["Data"]
	if !found {
		return nil, fmt.Errorf("data missing in item: %v", i.Item)
	}

	err = json.Unmarshal([]byte(*data.S), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *DynamoDBCertStore) putItem(item *dynamodb.PutItemInput) error {
	item.TableName = aws.String(s.tableName)
	_, err := s.dyndbSvc.PutItem(item)
	return err
}

func (s *DynamoDBCertStore) putCert(crt *dynamoDBCert) error {
	item := &dynamodb.PutItemInput{}

	now := time.Now().UTC()
	if crt.Version == 0 {
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

		// check that the versions match
		item.ConditionExpression = aws.String("Version = :version")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":version": &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(crt.Version))},
		}
	}

	crt.Version++

	data, err := json.Marshal(crt)
	if err != nil {
		return err
	}

	item.Item = map[string]*dynamodb.AttributeValue{
		"Hash":     &dynamodb.AttributeValue{S: aws.String(hash(crt.Domains))},
		"Version":  &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(crt.Version))},
		"Data":     &dynamodb.AttributeValue{S: aws.String(string(data))},
		"NotAfter": &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(crt.NotAfter.UTC().Unix(), 10))},
	}

	return s.putItem(item)
}

// reserveCert tries to reserve a non-existing certificate in the store.
// Returns errUnableToReserveCertificate if it cannot reserve the certificate.
func (s *DynamoDBCertStore) reserveCert(domains []string) (*dynamoDBCert, error) {
	c := &dynamoDBCert{
		Domains: domains,
		State:   dynamoDBCertStateReserved,
		// set NotAfter to a high value to prevent cleaning by the DynamoDB TTL process
		NotAfter: time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	err := s.putCert(c)
	if err != nil {
		if awsErr, isAws := err.(awserr.Error); isAws {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return nil, errUnableToReserveCertificate
			}
		}

		return nil, err
	}

	return c, nil
}

// updateCert tries to update an existing certificate in the store.
// Returns errUnableToUpdateCertificate if it cannot update the certificate.
func (s *DynamoDBCertStore) updateCert(crt *dynamoDBCert) error {
	crt.Modified = time.Now().UTC()

	err := s.putCert(crt)
	if err != nil {
		if awsErr, isAws := err.(awserr.Error); isAws {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return errUnableToUpdateCertificate
			}
		}

		return err
	}

	return nil
}

// Save tries to save a certificate to the store. It is concurrent.
// Providing a nil certificate creates a reservation in the store.
func (s *DynamoDBCertStore) Save(domains []string, crt *certs.Certificate) error {
	if crt == nil {
		// make a reservation for this certificate
		_, err := s.reserveCert(domains)
		if err != nil && err == errUnableToReserveCertificate {
			// this means that another routine created a reservation first, that's ok
			return nil
		}

		return err
	}

	// check if there is an existing certificate in the store
	c, err := s.getCert(domains)
	if err != nil {
		return nil
	}

	if c == nil {
		// create new certificate
		c = &dynamoDBCert{
			Domains: domains,
		}
	}

	// we've received a non-nil certificate so update state to generated and
	// copy the private key and certificate
	c.State = dynamoDBCertStateGenerated
	c.PrivateKey = string(crt.PrivateKey)
	c.Certificate = string(crt.Certificate)

	// get expiry date from certificate
	c.NotAfter, err = commonCerts.NotAfter(crt)
	if err != nil {
		return err
	}

	return s.putCert(c)
}
