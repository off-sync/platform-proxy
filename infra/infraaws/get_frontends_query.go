package infraaws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/off-sync/platform-proxy/app/frontends/qry/getfrontends"
	"github.com/off-sync/platform-proxy/domain/frontends"
)

type DynamoDBGetFrontendsQuery struct {
	dyndbSvc  *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBGetFrontendsQuery(p client.ConfigProvider, tableName string) (*DynamoDBGetFrontendsQuery, error) {
	dyndbSvc := dynamodb.New(p)

	_, err := dyndbSvc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	return &DynamoDBGetFrontendsQuery{
		dyndbSvc:  dyndbSvc,
		tableName: tableName,
	}, nil
}

type frontendItem struct {
	DomainName           string                     `dynamodbav:"DomainName"`
	Certificate          string                     `dynamodbav:"Certificate"`
	PrivateKey           string                     `dynamodbav:"PrivateKey"`
	CertificateExpiresAt dynamodbattribute.UnixTime `dynamodbav:"CertificateExpiresAt"`
	ServiceName          string                     `dynamodbav:"ServiceName"`
}

func (q *DynamoDBGetFrontendsQuery) Execute(model *getfrontends.QueryModel) (*getfrontends.ResultModel, error) {
	result := &getfrontends.ResultModel{
		Frontends: []*frontends.Frontend{},
	}

	var unmarshalErr error

	err := q.dyndbSvc.ScanPages(&dynamodb.ScanInput{
		TableName: aws.String(q.tableName),
	}, func(page *dynamodb.ScanOutput, last bool) bool {
		frontendItems := []*frontendItem{}

		err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &frontendItems)
		if err != nil {
			unmarshalErr = err

			return false
		}

		for _, i := range frontendItems {
			result.Frontends = append(result.Frontends, &frontends.Frontend{
				DomainName:           i.DomainName,
				Certificate:          i.Certificate,
				PrivateKey:           i.PrivateKey,
				CertificateExpiresAt: time.Time(i.CertificateExpiresAt),
				ServiceName:          i.ServiceName,
			})
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return result, nil
}
