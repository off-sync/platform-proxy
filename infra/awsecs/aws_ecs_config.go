package awsecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	serverContainerName = "server"
	dockerLabelPort     = "com.off-sync.platform.proxy.port"
	defaultPort         = 8080
)

// ConfigProvider provides an AWS ECS based ConfigProvider implementation.
type ConfigProvider struct {
	notificationChans []chan<- bool
	ecsSvc            *ecs.ECS
	cluster           *ecs.Cluster
}

// New returns a new AWS ECS Configuration Provider. It checks the cluster
// before returning.
func New(ecsSvc *ecs.ECS, clusterName string) (*ConfigProvider, error) {
	clusters, err := ecsSvc.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	})
	if err != nil {
		return nil, err
	}

	if len(clusters.Failures) > 0 {
		return nil, fmt.Errorf("checking cluster: %s", *clusters.Failures[0].Reason)
	}

	if len(clusters.Clusters) < 1 {
		return nil, fmt.Errorf("cluster not found")
	}

	return &ConfigProvider{
		ecsSvc:  ecsSvc,
		cluster: clusters.Clusters[0],
	}, nil
}

// GetNotificationChannel creates a new channel to which configuration updates are sent.
func (p *ConfigProvider) GetNotificationChannel() chan<- bool {
	c := make(chan<- bool, 1)
	p.notificationChans = append(p.notificationChans, c)

	return c
}
