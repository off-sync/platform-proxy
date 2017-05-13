package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/off-sync/platform-proxy/infra/infraaws"
)

var log = logrus.New()

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	getFrontends, err := infraaws.NewDynamoDBGetFrontendsQuery(sess, "off-sync-qa-frontends")
	if err != nil {
		log.WithError(err).Fatal("creating DynamoDB GetFrontends query")
	}

	frontends, err := getFrontends.Execute(nil)
	if err != nil {
		log.WithError(err).Fatal("getting frontends")
	}

	for _, frontend := range frontends.Frontends {
		log.
			WithField("domain_name", frontend.DomainName).
			WithField("service_name", frontend.ServiceName).
			Info("frontend")
	}

	getServices, err := infraaws.NewEcsGetServicesQuery(sess, "off-sync-qa")
	if err != nil {
		log.WithError(err).Fatal("creating ECS GetServices query")
	}

	services, err := getServices.Execute(nil)
	if err != nil {
		log.WithError(err).Fatal("getting services")
	}

	for _, service := range services.Services {
		log.
			WithField("name", service.Name).
			WithField("servers", service.Servers).
			Info("service")
	}
}
