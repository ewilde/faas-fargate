package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/openfaas/faas/gateway/requests"
)

func CreateTaskRevision(client *ecs.ECS, request requests.CreateFunctionRequest) (*ecs.RegisterTaskDefinitionOutput, error) {
	registration, err := client.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family: aws.String(ServiceNameFromFunctionName(request.Service)),
		Memory: aws.String("512"), // TODO configure
		Cpu:    aws.String("256"), // TODO configure
		RequiresCompatibilities: []*string{aws.String("FARGATE")},
		NetworkMode:             aws.String("awsvpc"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:  aws.String(ServiceNameFromFunctionName(request.Service)),
				Image: aws.String(request.Image),
			},
		},

	})

	if err != nil {
		return nil, err
	}

	return registration, nil
}
