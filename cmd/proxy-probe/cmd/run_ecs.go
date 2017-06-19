package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/viper"
)

var (
	ecsClusterName string
)

func init() {
	runCmd.PersistentFlags().StringVarP(&ecsClusterName, "ecs-cluster-name", "c", "", "ECS cluster containing the backends")
	viper.BindPFlag("ecsClusterName", runCmd.PersistentFlags().Lookup("ecs-cluster-name"))
}

func checkECS(p client.ConfigProvider) {
	ecsSvc := ecs.New(p)

	clusterName := viper.GetString("ecsClusterName")

	// ECS::DescribeClusters
	le := log.WithField("cluster_name", clusterName)

	clusters, err := ecsSvc.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{&clusterName},
	})
	if err != nil {
		le.WithError(err).Fatal("ECS::DescribeClusters failed")
	}

	if len(clusters.Clusters) != 1 {
		le.Fatal("ECS::DescribeClusters failed: cluster not found")
	}

	clusterArn := *clusters.Clusters[0].ClusterArn

	le.WithField("cluster_arn", clusterArn).Info("ECS::DescribeClusters successful")

	// ECS::ListServices
	le = log.WithField("cluster_arn", clusterArn)

	services, err := ecsSvc.ListServices(&ecs.ListServicesInput{
		Cluster:    &clusterName,
		MaxResults: aws.Int64(1),
	})
	if err != nil {
		le.WithError(err).Fatal("ECS::ListServices")
	}

	if len(services.ServiceArns) != 1 {
		le.Fatal("ECS::ListServices failed: no services found")
	}

	serviceArn := *services.ServiceArns[0]

	le.WithField("service_arn", serviceArn).Info("ECS::ListServices successful")

	// ECS::DescribeServices
	le = log.WithField("service_arn", serviceArn)

	clusterServices, err := ecsSvc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  &clusterArn,
		Services: []*string{&serviceArn},
	})
	if err != nil {
		le.WithError(err).Fatal("ECS::DescribeServices failed")
	}

	if len(clusterServices.Services) != 1 {
		le.WithError(err).Fatal("ECS::DescribeServices failed: service description not found")
	}

	taskDefArn := *clusterServices.Services[0].TaskDefinition

	le.WithField("task_def_arn", taskDefArn).Info("ECS::DescribeServices successful")

	// ECS::DescribeTaskDefinition
	le = log.WithField("task_def_arn", taskDefArn)

	tdef, err := ecsSvc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefArn,
	})
	if err != nil {
		le.WithError(err).Fatal("ECS::DescribeTaskDefinition failed")
	}

	le.WithField("family", *tdef.TaskDefinition.Family).Info("ECS::DescribeTaskDefinition successful")
}
