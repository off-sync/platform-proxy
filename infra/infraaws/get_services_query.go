package infraaws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/off-sync/platform-proxy/app/services/qry/getservices"
	"github.com/off-sync/platform-proxy/domain/services"
)

const (
	serverContainerName = "server"
	dockerLabelPort     = "com.off-sync.platform.proxy.port"
	defaultPort         = 8080
)

type EcsGetServicesQuery struct {
	ecsSvc  *ecs.ECS
	cluster *ecs.Cluster
}

func NewEcsGetServicesQuery(p client.ConfigProvider, clusterName string) (*EcsGetServicesQuery, error) {
	ecsSvc := ecs.New(p)

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

	return &EcsGetServicesQuery{
		ecsSvc:  ecsSvc,
		cluster: clusters.Clusters[0],
	}, nil
}

// Execute returns the services of a ECS cluster. It processes all services (list-services),
// describes them (describe-services), and describes the service's task definition (describe-task-definition).
// An item is returned for each service that has a container with the name 'server' in its task definition.
func (q *EcsGetServicesQuery) Execute(model *getservices.QueryModel) (*getservices.ResultModel, error) {
	var servicesErr error

	serviceArns := make(chan *string)
	go func() {
		servicesErr = q.getServiceArns(serviceArns)

		close(serviceArns)
	}()

	result := &getservices.ResultModel{
		Services: []*services.Service{},
	}

	for serviceArn := range serviceArns {
		clusterServices, err := q.ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  q.cluster.ClusterArn,
			Services: []*string{serviceArn},
		})
		if err != nil {
			return nil, err
		}

		service := clusterServices.Services[0]

		taskDefArn := service.TaskDefinition

		taskDefServer, err := q.getTaskDefinitionServer(taskDefArn)
		if err != nil {
			return nil, err
		}

		if taskDefServer == "" {
			// no server container present in this task definition
			continue
		}

		resultService, err := services.NewService(*service.ServiceName, taskDefServer)
		if err != nil {
			return nil, err
		}

		result.Services = append(result.Services, resultService)
	}

	if servicesErr != nil {
		return nil, servicesErr
	}

	return result, nil
}

func (q *EcsGetServicesQuery) getServiceArns(out chan<- *string) error {
	var nextToken *string
	for {
		services, err := q.ecsSvc.ListServices(&ecs.ListServicesInput{
			Cluster:   q.cluster.ClusterArn,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}

		for _, serviceArn := range services.ServiceArns {
			out <- serviceArn
		}

		nextToken = services.NextToken
		if nextToken == nil {
			break
		}
	}

	return nil
}

func (q *EcsGetServicesQuery) getTaskDefinitionServer(taskDefArn *string) (string, error) {
	tdef, err := q.ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskDefArn,
	})
	if err != nil {
		return "", err
	}

	for _, cdef := range tdef.TaskDefinition.ContainerDefinitions {
		if *cdef.Name != serverContainerName {
			// not the server
			continue
		}

		port := defaultPort

		portLabel, found := cdef.DockerLabels[dockerLabelPort]
		if found {
			port, err = strconv.Atoi(*portLabel)
			if err != nil {
				return "", fmt.Errorf("invalid port: %s", *portLabel)
			}
		}

		return fmt.Sprintf("http://%s:%d", *cdef.Hostname, port), nil
	}

	return "", nil
}
