package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func buildLogPolicyStatement(
	builder *PolicyBuilder,
	name string) error {

	builder.AddStatement(
		[]string{
			"logs:CreateLogStream",
			"logs:PutLogEvents",
		},
		[]string{
			fmt.Sprintf("arn:aws:logs:*:%s:*", name),
		})

	return nil
}

func createLogGroup(functionName string) (string, error) {
	name := ServiceNameFromFunctionName(functionName)
	_, err := cloudwatchClient.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
				return name, nil
			}
		}

		return "", fmt.Errorf("error creating log group for %s. %v", functionName, err)
	}

	return name, nil
}

func deleteLogGroup(functionName string) error {
	name := ServiceNameFromFunctionName(functionName)
	_, err := cloudwatchClient.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error deleting log group for %s. %v", functionName, err)
	}

	return nil
}
