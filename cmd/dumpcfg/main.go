package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"sync"

	"github.com/off-sync/platform-proxy/app/config/cmd/startwatcher"
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

	startWatcher, err := infraaws.NewSqsStartWatcherCommand(sess, "off-sync-qa-platform-proxy-config", 5)
	if err != nil {
		log.WithError(err).Fatal("creating SQS StartWatcher command")
	}

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())

	startWatcher.Execute(&startwatcher.CommandModel{
		WaitGroup: wg,
		Ctx:       ctx,
		Callback: func(changes *startwatcher.Changes) {
			log.
				WithField("services", changes.Services).
				WithField("frontends", changes.Frontends).
				Info("received change notification")
		},
	})

	// create signal channel and wait for SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs

	log.WithField("signal", sig).Info("exiting because of received signal")

	// cancel context
	cancel()

	// wait for other routines to complete
	wg.Wait()
}
