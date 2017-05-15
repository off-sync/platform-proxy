package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/off-sync/platform-proxy/app/config/cmd/startwatcher"
	"github.com/off-sync/platform-proxy/app/frontends/qry/getfrontends"
	"github.com/off-sync/platform-proxy/app/services/qry/getservices"
	"github.com/off-sync/platform-proxy/infra/infraaws"
)

var getFrontendsQuery getfrontends.Query
var getServicesQuery getservices.Query
var startWatcherCommand startwatcher.Command

func init() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	getFrontendsQuery, err = infraaws.NewDynamoDBGetFrontendsQuery(sess, "off-sync-qa-frontends")
	if err != nil {
		log.WithError(err).Fatal("creating DynamoDB GetFrontends query")
	}

	getServicesQuery, err = infraaws.NewEcsGetServicesQuery(sess, "off-sync-qa")
	if err != nil {
		log.WithError(err).Fatal("creating ECS GetServices query")
	}

	startWatcherCommand, err = infraaws.NewSqsStartWatcherCommand(sess, "off-sync-qa-platform-proxy-config", 5)
	if err != nil {
		log.WithError(err).Fatal("creating SQS StartWatcher command")
	}
}
