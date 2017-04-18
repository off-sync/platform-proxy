package dyndbutil

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func StringAttr(s string) *dynamodb.AttributeValue {
	if s == "" {
		return &dynamodb.AttributeValue{NULL: aws.Bool(true)}
	}

	return &dynamodb.AttributeValue{S: aws.String(s)}
}

func StringValue(a *dynamodb.AttributeValue) string {
	if aws.BoolValue(a.NULL) {
		return ""
	}

	return aws.StringValue(a.S)
}

func StringListAttr(s []string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{SS: aws.StringSlice(s)}
}

func TimeAttr(t time.Time) *dynamodb.AttributeValue {
	if t.IsZero() {
		return &dynamodb.AttributeValue{NULL: aws.Bool(true)}
	}

	return &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(t.UTC().Unix(), 10))}
}

func TimeValue(a *dynamodb.AttributeValue) (time.Time, error) {
	if aws.BoolValue(a.NULL) {
		return time.Time{}, nil
	}

	s, err := strconv.ParseInt(aws.StringValue(a.N), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(s, 0), nil
}
