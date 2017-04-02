package awsecs

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/off-sync/platform-proxy/domain/sites"
)

// GetBackends returns the backends of a ECS cluster. It processes all service (list-services),
// describes them (describe-services), and describes the service's task definition (describe-task-definition).
// A backend is returned for each service that has a container with the name 'server' in its task definition.
func (p *ConfigProvider) GetBackends() ([]*sites.Backend, error) {
	var servicesErr error

	serviceArns := make(chan *string)
	go func() {
		servicesErr = p.getServiceArns(serviceArns)

		close(serviceArns)
	}()

	var backends []*sites.Backend

	for serviceArn := range serviceArns {
		services, err := p.ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  p.cluster.ClusterArn,
			Services: []*string{serviceArn},
		})
		if err != nil {
			return nil, err
		}

		service := services.Services[0]

		taskDefArn := service.TaskDefinition

		taskDefServer, err := p.getTaskDefinitionServer(taskDefArn)
		if err != nil {
			return nil, err
		}

		if taskDefServer == "" {
			// no server container present in this task definition
			continue
		}

		backend, err := sites.NewBackend(*service.ServiceName, taskDefServer)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	if servicesErr != nil {
		return nil, servicesErr
	}

	return backends, nil
}

func (p *ConfigProvider) getServiceArns(out chan<- *string) error {
	var nextToken *string
	for {
		services, err := p.ecsSvc.ListServices(&ecs.ListServicesInput{
			Cluster:   p.cluster.ClusterArn,
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

func (p *ConfigProvider) getTaskDefinitionServer(taskDefArn *string) (string, error) {
	tdef, err := p.ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
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
