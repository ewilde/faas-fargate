package aws

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/ewilde/faas-ecs/types"
	"github.com/ewilde/faas-ecs/system"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

const servicePrefix = "openfaas-"

var clusterID string
var subnetsFunc = &sync.Once{}
var subnets []*string

func init() {
	clusterID = system.GetEnv("cluster_name", "openfaas")
}

func FindECSServiceArn(client *ecs.ECS, serviceName string) (*string, error) {
	services, err := client.ListServices(&ecs.ListServicesInput{
		Cluster: ClusterID(),
	})

	if err != nil {
		return nil, err
	}

	var serviceArn *string = nil
	for _, item := range services.ServiceArns {
		if strings.Contains(aws.StringValue(item), serviceName) {
			serviceArn = item
		}
	}

	return serviceArn, nil
}

func UpdateOrCreateECSService(
	ecsClient *ecs.ECS,
	ec2Client *ec2.EC2,
	discovery *servicediscovery.ServiceDiscovery,
	taskDefinition *ecs.TaskDefinition,
	request requests.CreateFunctionRequest,
	cfg *types.DeployHandlerConfig) (*ecs.Service, error) {

	serviceArn, err := FindECSServiceArn(ecsClient, request.Service)
	if err != nil {
		log.Errorln(fmt.Sprintf("Could not find service with name %s.", request.Service), err)
		return nil, err
	}

	if serviceArn != nil {
		service, err := ecsClient.UpdateService(&ecs.UpdateServiceInput{
			Cluster:        ClusterID(),
			Service:        serviceArn,
			DesiredCount:   getMinReplicaCount(request.Labels),
			TaskDefinition: taskDefinition.TaskDefinitionArn,
		})

		if err != nil {
			log.Errorln(fmt.Sprintf("Error updating service %s. ", request.Service), err)
			return nil, err
		}

		return service.Service, err
	}

	registryArn, err := ensureServiceRegistrationExists(discovery, request.Service, cfg.VpcID)
	if err != nil {
		log.Errorln(fmt.Sprintf("Error creating registry for service %s. ", request.Service), err)
		return nil, err
	}

	// see: https://docs.aws.amazon.com/cli/latest/reference/ecs/create-service.html
	result, err := ecsClient.CreateService(&ecs.CreateServiceInput{
		Cluster:        ClusterID(),
		ServiceName:    aws.String(ServiceNameFromFunctionName(request.Service)),
		TaskDefinition: taskDefinition.TaskDefinitionArn,
		LaunchType:     aws.String("FARGATE"),
		DesiredCount:   getMinReplicaCount(request.Labels),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(cfg.AssignPublicIP),
				Subnets:        awsSubnet(ec2Client, cfg.SubnetIDs, cfg.VpcID),
				SecurityGroups: []*string { aws.String(cfg.SecurityGroupId) },
			},
		},
		ServiceRegistries: []*ecs.ServiceRegistry{
			{
				RegistryArn: aws.String(registryArn),
			},
		},
	})

	if err != nil {
		log.Errorln(fmt.Sprintf("Error creating service %s. Using subnets from configuration: %s",
			request.Service, cfg.SubnetIDs), err)
		return nil, err
	}

	return result.Service, nil
}

func UpdateECSServiceDesiredCount(
	ecsClient *ecs.ECS,
	serviceName string,
	desiredCount int) (*ecs.Service, error) {

	serviceArn, err := FindECSServiceArn(ecsClient, serviceName)
	if err != nil {
		log.Errorln(fmt.Sprintf("Could not find service with name %s.", serviceName), err)
		return nil, err
	}

	if serviceArn == nil {
		return nil, errors.New(fmt.Sprintf("Could not find service %s", serviceName))
	}

	service, err := ecsClient.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      ClusterID(),
		Service:      serviceArn,
		DesiredCount: aws.Int64(int64(desiredCount)),
	})

	if err != nil {
		return nil, err
	}

	return service.Service, nil
}

func ClusterID() *string {
	return aws.String(clusterID)
}

func IsFaasService(arn *string) bool {
	return strings.Contains(aws.StringValue(arn), servicePrefix)
}

func ServiceNameFromArn(arn *string) *string {
	return aws.String(strings.Split(*arn, "/")[1])
}

func ServiceNameForDisplay(name *string) string {
	return strings.TrimPrefix(*name, servicePrefix)
}

func ServiceNameFromFunctionName(functionName string) string {
	return servicePrefix + functionName
}

func awsSubnet(client *ec2.EC2, subnetIds string, vpcId string) []*string {

	subnetsFunc.Do(func() {
		if subnetIds == "" {
			log.Debugf("Searching for subnets using vpc id %s", vpcId)
			result, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("vpc-id"),
						Values: []*string{
							aws.String(vpcId),
						},
					},
				},
			})
			if err == nil {
				for _, item := range result.Subnets {
					subnets = append(subnets, item.SubnetId)
				}
			}
		} else {
			log.Debugf("Searching for subnets using list of ids %s", subnetIds)
			subnetIds :=strings.Split(subnetIds, ",")
			for _, item := range subnetIds {
				subnets = append(subnets, aws.String(item))
			}
		}
	})

	return subnets
}

func getMinReplicaCount(labels *map[string]string) *int64 {
	if labels == nil {
		return aws.Int64(1)
	}

	m := *labels
	if value, exists := m["com.openfaas.scale.min"]; exists {
		minReplicas, err := strconv.Atoi(value)
		if err == nil && minReplicas > 0 {
			return aws.Int64(int64(minReplicas))
		}

		log.Errorln(err)
	}

	return aws.Int64(1)
}
