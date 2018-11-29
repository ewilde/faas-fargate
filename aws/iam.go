package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

// PolicyBuilder is used to build IAM policies
type PolicyBuilder struct {
	statements []string
}

// NewPolicyBuilder create a new PolicyBuilder
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{}
}

// AddStatement adds a statement with corresponding actions and resources
func (p *PolicyBuilder) AddStatement(actions []string, resources []string) {
	p.statements = append(p.statements, fmt.Sprintf(statementTemplate, p.normalizeStrings(actions), p.normalizeStrings(resources)))
}

func (p *PolicyBuilder) normalizeStrings(items []string) string {
	builder := strings.Builder{}
	for _, item := range items {
		builder.WriteString(fmt.Sprintf("\"%s\",", item))
	}

	return strings.TrimRight(builder.String(), ",")
}

func (p *PolicyBuilder) normalizeObjects(items []string) string {
	return strings.TrimRight(strings.Join(items, ","), ",")
}

func (p *PolicyBuilder) String() string {
	return fmt.Sprintf(policyTemplate, p.normalizeObjects(p.statements))
}

const policyTemplate = `{
  "Version": "2012-10-17",
  "Statement": [%s]
}`

const statementTemplate = `{
        "Effect": "Allow",
        "Action": [%s],
        "Resource": [%s]
    }`

func createRoleWithPolicy(functionName string, policyDocument string) (string, error) {
	roleName := ServiceNameFromFunctionName(functionName)

	existing, err := iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if checkForErrorAllowEntityNotExists(err) != nil {
		return "", err
	}

	var roleArn *string
	if existing.Role == nil {
		output, err := iamClient.CreateRole(&iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		})

		if err != nil {
			return "", fmt.Errorf("could not create role %s. %v", roleName, err)
		}

		roleArn = output.Role.Arn
	} else {
		roleArn = existing.Role.Arn
	}

	_, err = iamClient.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(fmt.Sprintf("%s-policy", roleName)),
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		return "", fmt.Errorf("could not create role policy %s\n%s\n. %v", roleName, policyDocument, err)
	}

	return aws.StringValue(roleArn), nil
}

func deleteRole(name string) error {
	roleName := ServiceNameFromFunctionName(name)

	_, err := iamClient.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		PolicyName: aws.String(fmt.Sprintf("%s-policy", roleName)),
		RoleName:   aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("could not delete role policy for role %s. %v", roleName, err)
	}

	_, err = iamClient.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("could not delete role %s. %v", roleName, err)
	}

	return nil
}

func checkForErrorAllowEntityNotExists(err error) error {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != iam.ErrCodeNoSuchEntityException {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

const assumeRolePolicy string = `{
  "Version": "2008-10-17",
  "Statement": [{
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "ecs.amazonaws.com",
          "ecs-tasks.amazonaws.com"
        ]
      },
      "Effect": "Allow"
    }]
}`
