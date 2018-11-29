package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

var (
	cloudwatchClient *cloudwatchlogs.CloudWatchLogs
	ecsClient        *ecs.ECS
	ec2Client        *ec2.EC2
	iamClient        *iam.IAM
	secretsClient    *secretsmanager.SecretsManager
	discoveryClient  *servicediscovery.ServiceDiscovery
)

func init() {
	session := session.Must(session.NewSession())
	logLevel := awsLogLevel()

	cloudwatchClient = cloudwatchlogs.New(session, aws.NewConfig().WithLogLevel(logLevel))
	ecsClient = ecs.New(session, aws.NewConfig().WithLogLevel(logLevel))
	ec2Client = ec2.New(session, aws.NewConfig().WithLogLevel(logLevel))
	iamClient = iam.New(session, aws.NewConfig().WithLogLevel(logLevel))
	secretsClient = secretsmanager.New(session, aws.NewConfig().WithLogLevel(logLevel))
	discoveryClient = servicediscovery.New(session, aws.NewConfig().WithLogLevel(logLevel))
}

// KeyValuePairGetValue searches the array of values and returns the matching name or nil if none are found.
func KeyValuePairGetValue(name string, values []*ecs.KeyValuePair) (*string, bool) {
	for _, item := range values {
		if *item.Name == name {
			return item.Value, true
		}
	}

	return nil, false
}

func awsLogLevel() aws.LogLevelType {
	lvl := os.Getenv("LOG_LEVEL")

	if lvl == "trace" {
		return aws.LogDebugWithRequestErrors
	}

	return aws.LogOff
}
