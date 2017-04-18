package acmestore

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/off-sync/platform-proxy/common/dyndbutil"
	"github.com/off-sync/platform-proxy/domain/acme"
	lego "github.com/xenolf/lego/acme"
)

// DynamoDBACMEStore implements the ACMESaver and ACMELoader interfaces
// using an AWS DynamoDB table as its backend.
// It is not concurrent-safe.
type DynamoDBACMEStore struct {
	dyndbSvc  *dynamodb.DynamoDB
	tableName string
}

func (s *DynamoDBACMEStore) getItem(accountKey string, attrs ...string) (*dynamodb.GetItemOutput, error) {
	return s.dyndbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountKey": dyndbutil.StringAttr(accountKey),
		},
		AttributesToGet: aws.StringSlice(attrs),
	})
}

func (s *DynamoDBACMEStore) putItem(item *dynamodb.PutItemInput) error {
	item.TableName = aws.String(s.tableName)
	_, err := s.dyndbSvc.PutItem(item)
	return err
}

func accountKey(endpoint, email string) string {
	return endpoint + "#" + email
}

// Load tries to load an ACME account from the DynamoDB store.
// It returns nil if the account does not exist.
func (s *DynamoDBACMEStore) Load(endpoint, email string) (*acme.Account, error) {
	i, err := s.getItem(accountKey(endpoint, email), "PrivateKey", "Registration")
	if err != nil {
		return nil, err
	}

	if i.Item == nil {
		// item not found
		return nil, nil
	}

	account := &acme.Account{
		Endpoint:     endpoint,
		Email:        email,
		PrivateKey:   dyndbutil.StringValue(i.Item["PrivateKey"]),
		Registration: &lego.RegistrationResource{},
	}

	err = json.Unmarshal([]byte(dyndbutil.StringValue(i.Item["Registration"])), account.Registration)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Save saves an ACME account to the DynamoDB store.
// The registration is marshalled to JSON and saved to a single field.
func (s *DynamoDBACMEStore) Save(account *acme.Account) error {
	reg, err := json.Marshal(account.Registration)
	if err != nil {
		return err
	}

	item := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"AccountKey":   dyndbutil.StringAttr(accountKey(account.Endpoint, account.Email)),
			"Endpoint":     dyndbutil.StringAttr(account.Endpoint),
			"Email":        dyndbutil.StringAttr(account.Email),
			"PrivateKey":   dyndbutil.StringAttr(account.PrivateKey),
			"Registration": dyndbutil.StringAttr(string(reg)),
		},
	}

	return s.putItem(item)
}
