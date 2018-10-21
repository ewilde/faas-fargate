package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// VpcFromSubnet calculates the vpc id based on the supplied subnet ids, or returns the default vpc id if subnets
// are an empty string
func VpcFromSubnet(client *ec2.EC2, subnets string) string {

	if subnets == "" {
		vpcResult, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{})
		if err != nil {
			log.Errorln("Error describing vpc. ", err)
			return ""
		}

		var defaultVpc *string
		for _, item := range vpcResult.Vpcs {
			if aws.BoolValue(item.IsDefault) {
				defaultVpc = item.VpcId
				break
			}
		}

		return aws.StringValue(defaultVpc)
	}

	subnetIds := strings.Split(subnets, ",")
	result, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(subnetIds[0])},
	})

	if err != nil {
		log.Errorln(fmt.Sprintf("Error describing subnet %s. ", subnetIds[0]), err)
		return ""
	}

	return aws.StringValue(result.Subnets[0].VpcId)
}
