package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy/infra/awsecs"
)

var log = logrus.New()

func main() {
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	sess, err := session.NewSession()
	if err != nil {
		log.WithError(err).Fatal("creating new session")
	}

	ecsSvc := ecs.New(sess, &aws.Config{Region: aws.String("eu-west-1")})

	p, err := awsecs.New(ecsSvc, "off-sync-qa")
	if err != nil {
		log.WithError(err).Fatal("creating AWS ECS config provider")
	}

	sites, err := p.GetSites()
	if err != nil {
		log.WithError(err).Fatal("getting sites")
	}

	for _, site := range sites {
		log.
			WithField("domains", site.Domains).
			WithField("backends", site.Backends).
			Info("site")
	}
}
