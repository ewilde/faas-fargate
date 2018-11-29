package aws

import (
	"fmt"
	"strings"

	"github.com/ewilde/faas-fargate/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/openfaas/faas/gateway/requests"
)

// CreateTaskRevision create a new task revision
func CreateTaskRevision(
	request requests.CreateFunctionRequest,
	config *types.DeployHandlerConfig) (*ecs.RegisterTaskDefinitionOutput, error) {

	totalMemory := 512
	totalCPU := 256
	funcMemory := totalMemory
	funcCPU := totalCPU

	name := ServiceNameFromFunctionName(request.Service)
	taskDefinitionInput := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(name),
		Memory: aws.String("512"), // TODO configure
		Cpu:    aws.String("256"), // TODO configure
		RequiresCompatibilities: []*string{aws.String("FARGATE")},
		NetworkMode:             aws.String("awsvpc"),
	}

	logGroupName, err := createLogGroup(request.Service)
	if err != nil {
		return nil, err
	}

	funcTask := &ecs.ContainerDefinition{
		Name:  aws.String(name),
		Image: aws.String(request.Image),
		LogConfiguration: &ecs.LogConfiguration{
			LogDriver: aws.String("awslogs"),
			Options: map[string]*string{
				"awslogs-group":         aws.String(logGroupName),
				"awslogs-region":        aws.String(config.Region),
				"awslogs-stream-prefix": aws.String(logGroupName),
			},
		},
	}

	policy := NewPolicyBuilder()
	err = buildLogPolicyStatement(policy, logGroupName)
	if err != nil {
		return nil, err
	}

	if len(request.Secrets) > 0 {
		err := buildSecretsPolicyStatement(policy, request.Service, request.Secrets)
		if err != nil {
			return nil, err
		}

		funcMemory = funcMemory - 32
		funcCPU = funcCPU - 64
		secretTask := &ecs.ContainerDefinition{
			Name:   aws.String(fmt.Sprintf("%s-kms", name)),
			Cpu:    aws.Int64(64),
			Memory: aws.Int64(32),
			Image:  aws.String("ewilde/kms-template:latest"),
			LogConfiguration: &ecs.LogConfiguration{
				LogDriver: aws.String("awslogs"),
				Options: map[string]*string{
					"awslogs-group":         aws.String(logGroupName),
					"awslogs-region":        aws.String(config.Region),
					"awslogs-stream-prefix": aws.String(fmt.Sprintf("%s-kms-template", logGroupName)),
				},
			},
		}

		secretTask.Environment = []*ecs.KeyValuePair{
			{
				Name:  aws.String("SECRETS"),
				Value: aws.String(strings.Join(getSecretNames(request.Secrets), ",")),
			},
		}

		taskDefinitionInput.ContainerDefinitions = append(taskDefinitionInput.ContainerDefinitions, secretTask)
		funcTask.VolumesFrom = []*ecs.VolumeFrom{{SourceContainer: secretTask.Name}}
	}

	funcTask.Cpu = aws.Int64(int64(funcCPU))
	funcTask.Memory = aws.Int64(int64(funcMemory))

	taskDefinitionInput.ContainerDefinitions = append(taskDefinitionInput.ContainerDefinitions, funcTask)

	arn, err := createRoleWithPolicy(request.Service, policy.String())
	if err != nil {
		return nil, err
	}

	taskDefinitionInput.TaskRoleArn = aws.String(arn)
	taskDefinitionInput.ExecutionRoleArn = aws.String(arn)

	output, err := ecsClient.RegisterTaskDefinition(taskDefinitionInput)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetLatestTaskRevision gets the latest task revision for the corresponding functionName
func GetLatestTaskRevision(functionName string) (string, error) {
	name := ServiceNameFromFunctionName(functionName)

	output, err := ecsClient.ListTaskDefinitions(&ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(name),
		Sort:         aws.String("DESC"),
	})

	if err != nil {
		return "", fmt.Errorf("error getting latest task revision for %s. %v", name, err)
	}

	return aws.StringValue(output.TaskDefinitionArns[0]), nil
}

// DeleteTaskRevision deletes the task revision
func DeleteTaskRevision(functionName string) error {
	latestTaskArn, err := GetLatestTaskRevision(functionName)
	if err != nil {
		return err
	}

	_, err = ecsClient.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(latestTaskArn),
	})

	if err != nil {
		return fmt.Errorf("error deleting task definition %s arn: %s. %v", functionName, latestTaskArn, err)
	}

	err = deleteRole(functionName)
	if err != nil {
		return fmt.Errorf("error deleting role for task definition %s arn: %s. %v", functionName, latestTaskArn, err)
	}

	err = deleteLogGroup(functionName)
	if err != nil {
		return fmt.Errorf("error deleting log group for task definition %s arn: %s. %v", functionName, latestTaskArn, err)
	}

	return err
}

func getSecretNames(secrets []string) []string {
	var names []string
	for _, v := range secrets {
		names = append(names, fmt.Sprintf("%s%s", servicePrefix, v))
	}

	return names
}
