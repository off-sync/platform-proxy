package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy/app/config/qry/getconfig"
	"github.com/off-sync/platform-proxy/infra/awsecs"
)

var getConfigQry *getconfig.Qry

func init() {
	sess, err := session.NewSession()
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	ecsSvc := ecs.New(sess, &aws.Config{Region: aws.String("eu-west-1")})

	provider, err := awsecs.New(ecsSvc, "off-sync-qa")
	if err != nil {
		log.WithError(err).Fatal("creating AWS ECS config provider")
	}

	getConfigQry = getconfig.New(provider)
}
