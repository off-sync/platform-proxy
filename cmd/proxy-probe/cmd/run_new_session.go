package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
)

var (
	awsRegion string
	awsID     string
	awsSecret string
)

func init() {
	runCmd.PersistentFlags().StringVarP(&awsID, "aws-region", "r", "", "AWS region to be used")
	viper.BindPFlag("awsRegion", runCmd.PersistentFlags().Lookup("aws-region"))

	runCmd.PersistentFlags().StringVarP(&awsID, "aws-id", "i", "", "AWS ID to be used")
	viper.BindPFlag("awsID", runCmd.PersistentFlags().Lookup("aws-id"))

	runCmd.PersistentFlags().StringVarP(&awsSecret, "aws-secret", "s", "", "AWS secret to be used")
	viper.BindPFlag("awsSecret", runCmd.PersistentFlags().Lookup("aws-secret"))
}

func newSession() client.ConfigProvider {
	awsRegion := viper.GetString("awsRegion")
	awsID := viper.GetString("awsID")
	awsSecret := viper.GetString("awsSecret")

	log.
		WithField("aws_region", awsRegion).
		WithField("aws_id", awsID).
		Info("running proxy probe")

	sess, err := session.NewSession(&aws.Config{
		Region:      &awsRegion,
		Credentials: credentials.NewStaticCredentials(awsID, awsSecret, ""),
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			log.Infof("AWS SDK", args)
		}),
	})
	if err != nil {
		log.WithError(err).Fatal("NewSession failed")
	}

	log.
		Info("NewSession successful")

	return sess
}
