package main

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy/app/interfaces"
	"github.com/off-sync/platform-proxy/infra/awsecs"
)

var log = logrus.New()

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	ecsSvc := ecs.New(sess)

	var p interfaces.ConfigProvider
	p, err = awsecs.New(ecsSvc, "off-sync-qa")
	if err != nil {
		log.WithError(err).Fatal("creating AWS ECS config provider")
	}

	backends, err := p.GetBackends()
	if err != nil {
		log.WithError(err).Fatal("getting backends")
	}

	for _, backend := range backends {
		log.
			WithField("name", backend.Name).
			WithField("servers", backend.Servers).
			Info("backend configuration")

		for _, server := range backend.Servers {
			addrs, err := net.LookupHost(server.Hostname())
			if err != nil {
				log.
					WithField("server", server).
					WithError(err).
					Fatal("looking up server host")
			}

			log.
				WithField("server", server).
				WithField("addrs", addrs).
				Info("server hostname lookup successful")
		}
	}

	frontends, err := p.GetFrontends()
	if err != nil {
		log.WithError(err).Fatal("getting backends")
	}

	for _, frontend := range frontends {
		log.
			WithField("domain", frontend.Domain).
			WithField("backend_name", frontend.BackendName).
			Info("frontend configuration")
	}
}
