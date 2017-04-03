package certstore

import (
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"encoding/json"

	"encoding/hex"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	commonCerts "github.com/off-sync/platform-proxy/common/certs"
	"github.com/off-sync/platform-proxy/domain/certs"
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
	PrivateKey  []byte
	Certificate []byte
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

func (s *DynamoDBCertStore) getDynamoDBCert(domains []string) (*dynamoDBCert, error) {
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

	err = json.Unmarshal(data.B, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Save tries to save a certificate to the store. It is concurrent.
func (s *DynamoDBCertStore) Save(domains []string, crt *certs.Certificate) error {
	c, err := s.getDynamoDBCert(domains)
	if err != nil {
		return nil
	}

	if c == nil {
		// certificate not found
		c = &dynamoDBCert{
			Domains: domains,
		}
	}

	if crt == nil {
		// if no certificate is provided, this is stored as a reserved state.
		// this prevents race conditions when generating new certificates which can
		// take some time to complete.
		c.State = dynamoDBCertStateReserved
		c.PrivateKey = nil
		c.Certificate = nil

		// set NotAfter to a high value to prevent cleaning by the DynamoDB TTL process
		c.NotAfter = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	} else {
		c.State = dynamoDBCertStateGenerated
		c.PrivateKey = crt.PrivateKey
		c.Certificate = crt.Certificate

		// parse certificate to get expiry date
		tlsCrt, err := commonCerts.ConvertToTLS(crt)
		if err != nil {
			return err
		}

		if len(tlsCrt.Certificate) < 1 {
			return fmt.Errorf("no errors in certificate: %s", string(crt.Certificate))
		}

		asnCrt, err := x509.ParseCertificate(tlsCrt.Certificate[0])
		if err != nil {
			return err
		}

		c.NotAfter = asnCrt.NotAfter
	}

	var condExpr *string
	var exprAttrValues map[string]*dynamodb.AttributeValue

	now := time.Now().UTC()
	if c.Version == 0 {
		// never saved before: set created
		c.Created = now

		condExpr = aws.String("attribute_not_exists(Version)")
	} else {
		// already exists: set modified
		c.Modified = now

		condExpr = aws.String("Version = :version")
		exprAttrValues = map[string]*dynamodb.AttributeValue{
			":version": &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(c.Version))},
		}
	}

	c.Version++

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	item := map[string]*dynamodb.AttributeValue{
		"Hash":     &dynamodb.AttributeValue{S: aws.String(hash(domains))},
		"Version":  &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(c.Version))},
		"Data":     &dynamodb.AttributeValue{S: aws.String(string(data))},
		"NotAfter": &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(c.NotAfter.UTC().Unix(), 10))},
	}

	_, err = s.dyndbSvc.PutItem(&dynamodb.PutItemInput{
		TableName:                 aws.String(s.tableName),
		Item:                      item,
		ConditionExpression:       condExpr,
		ExpressionAttributeValues: exprAttrValues,
	})

	return err
}
