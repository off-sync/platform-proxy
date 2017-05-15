package infraaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/off-sync/platform-proxy/app/frontends/cmd/setcertificate"
)

type DynamodDBSetCertificateCommand struct {
	dyndbSvc  *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBSetCertificateCommand(p client.ConfigProvider, tableName string) (*DynamodDBSetCertificateCommand, error) {
	dyndbSvc := dynamodb.New(p)

	_, err := dyndbSvc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	return &DynamodDBSetCertificateCommand{
		dyndbSvc:  dyndbSvc,
		tableName: tableName,
	}, nil
}

type setFrontendCertificateKey struct {
	DomainName string
}

type setFrontendCertificateValues struct {
	Certificate          string                     `dynamodbav:":certificate"`
	PrivateKey           string                     `dynamodbav:":privateKey"`
	CertificateExpiresAt dynamodbattribute.UnixTime `dynamodbav:":certificateExpiresAt"`
}

func (c *DynamodDBSetCertificateCommand) Execute(model *setcertificate.CommandModel) error {
	key, err := dynamodbattribute.MarshalMap(&setFrontendCertificateKey{
		DomainName: model.DomainName,
	})
	if err != nil {
		return err
	}

	values, err := dynamodbattribute.MarshalMap(&setFrontendCertificateValues{
		Certificate:          model.Certificate,
		PrivateKey:           model.PrivateKey,
		CertificateExpiresAt: dynamodbattribute.UnixTime(model.CertificateExpiresAt),
	})
	if err != nil {
		return err
	}

	_, err = c.dyndbSvc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String(c.tableName),
		Key:                       key,
		UpdateExpression:          aws.String("SET Certificate = :certificate, PrivateKey = :privateKey, CertificateExpiresAt = :certificateExpiresAt"),
		ExpressionAttributeValues: values,
	})
	if err != nil {
		return err
	}

	return nil
}
