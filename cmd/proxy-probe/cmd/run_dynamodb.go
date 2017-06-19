package cmd

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/spf13/viper"
)

var (
	dyndbFrontendsTable string
)

func init() {
	runCmd.PersistentFlags().StringVarP(&dyndbFrontendsTable, "dyndb-frontends-table", "f", "", "DynamoDB table holding the frontends")
	viper.BindPFlag("dyndbFrontendsTable", runCmd.PersistentFlags().Lookup("dyndb-frontends-table"))
}

func checkDynamoDB(p client.ConfigProvider) {
	dyndb := dynamodb.New(p)

	dyndbFrontendsTable := viper.GetString("dyndbFrontendsTable")

	// DynamoDB::DescribeTable
	le := log.WithField("table_name", dyndbFrontendsTable)

	_, err := dyndb.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: &dyndbFrontendsTable,
	})
	if err != nil {
		le.
			WithError(err).
			Fatal("DynamoDB::DescribeTable failed")
	}

	le.Info("DynamoDB::DescribeTable successful")

	// DynamoDB::ScanPages
	le = log.WithField("table_name", dyndbFrontendsTable)

	err = dyndb.ScanPages(&dynamodb.ScanInput{
		TableName: &dyndbFrontendsTable,
	}, func(page *dynamodb.ScanOutput, last bool) bool {
		return false
	})
	if err != nil {
		le.
			WithError(err).
			Fatal("DynamoDB::ScanPages failed")
	}

	le.Info("DynamoDB::ScanPages successful")
}
