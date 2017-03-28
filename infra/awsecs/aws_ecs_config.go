package awsecs

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy/domain/sites"
)

const (
	DockerLabelDomains = "com.off-sync.platform.proxy.hostnames"
	DockerLabelPort    = "com.off-sync.platform.proxy.port"
	DefaultPort        = 8080
)

type AwsEcsConfigProvider struct {
	notificationChans []chan<- bool
	ecsSvc            *ecs.ECS
	cluster           *ecs.Cluster
}

// New returns a new AWS ECS Configuration Provider. It checks the cluster
// before returning.
func New(ecsSvc *ecs.ECS, clusterName string) (*AwsEcsConfigProvider, error) {
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

	return &AwsEcsConfigProvider{
		ecsSvc:  ecsSvc,
		cluster: clusters.Clusters[0],
	}, nil
}

func (p *AwsEcsConfigProvider) GetNotificationChannel() chan<- bool {
	c := make(chan<- bool, 1)
	p.notificationChans = append(p.notificationChans, c)

	return c
}

// TODO list-services > describe-services > deployments > task-defition > container-definition
func (p *AwsEcsConfigProvider) GetSites() ([]*sites.Site, error) {
	var sites []*sites.Site

	var tasksErr error
	taskArns := make(chan *string)
	go func() {
		tasksErr = p.getTasks(taskArns)

		close(taskArns)
	}()

	for taskArn := range taskArns {
		taskSites, err := p.getTaskSite(taskArn)
		if err != nil {
			return nil, err
		}

		if len(taskSites) == 0 {
			// no site required for this service
			continue
		}

		sites = append(sites, taskSites...)
	}

	if tasksErr != nil {
		return nil, tasksErr
	}

	return sites, nil
}

func (p *AwsEcsConfigProvider) getTasks(out chan<- *string) error {
	var nextToken *string
	for {
		tasks, err := p.ecsSvc.ListTasks(&ecs.ListTasksInput{
			Cluster:       p.cluster.ClusterArn,
			DesiredStatus: aws.String(ecs.DesiredStatusRunning),
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}

		for _, taskArn := range tasks.TaskArns {
			out <- taskArn
		}

		nextToken = tasks.NextToken
		if nextToken == nil {
			break
		}
	}

	return nil
}

func (p *AwsEcsConfigProvider) getTaskSite(taskArn *string) ([]*sites.Site, error) {
	tasks, err := p.ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: p.cluster.ClusterArn,
		Tasks:   []*string{taskArn},
	})
	if err != nil {
		return nil, err
	}

	if len(tasks.Tasks) < 1 {
		return nil, errors.New("task not found")
	}

	task := tasks.Tasks[0]

	tdef, err := p.ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	var tsites []*sites.Site

	for _, cdef := range tdef.TaskDefinition.ContainerDefinitions {
		domainsLabel, found := cdef.DockerLabels[DockerLabelDomains]
		if !found {
			continue
		}

		port := DefaultPort

		portLabel, found := cdef.DockerLabels[DockerLabelPort]
		if found {
			port, _ = strconv.Atoi(*portLabel)
			if port == 0 {
				return nil, fmt.Errorf("invalid port: %s", *portLabel)
			}
		}

		hostAddrs, err := net.LookupHost(*cdef.Hostname)
		if err != nil {
			return nil, err
		}

		backends := make([]string, len(hostAddrs))
		for i, hostAddr := range hostAddrs {
			backends[i] = fmt.Sprintf("http://%s:%d", hostAddr, port)
		}

		tsite, err := sites.New(strings.Split(*domainsLabel, ","), backends)
		if err != nil {
			return nil, err
		}

		tsites = append(tsites, tsite)
	}

	return tsites, nil
}
