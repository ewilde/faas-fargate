package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func buildSecretsPolicyStatement(
	builder *PolicyBuilder,
	function string,
	secretNames []string) error {
	ids, err := getSecretsID(secretNames)
	if err != nil {
		return fmt.Errorf("could not create secret ids for %s. %v", function, err)
	}

	builder.AddStatement([]string{"secretsmanager:GetSecretValue"}, ids)
	return nil
}

func getSecretsID(names []string) ([]string, error) {
	var result []string

	for _, v := range names {
		name := fmt.Sprintf("%s%s", servicePrefix, v)
		output, err := secretsClient.DescribeSecret(&secretsmanager.DescribeSecretInput{
			SecretId: aws.String(name),
		})

		if err != nil {
			return nil, fmt.Errorf("error describing secret %s. %v", name, err)
		}

		result = append(result, aws.StringValue(output.ARN))
	}

	return result, nil
}
