package aws

import (
	"os"
	"testing"

	"github.com/ewilde/faas-fargate/types"
	"github.com/openfaas/faas/gateway/requests"
)

func TestAccCreateTaskRevision(t *testing.T) {
	PreTest(t)
	subnetIDs := os.Getenv("subnet_ids")
	vpcID := VpcFromSubnet(subnetIDs)

	_, err := CreateTaskRevision(requests.CreateFunctionRequest{
		Service: "figlet",
		Image:   "functions/figlet",
	}, &types.DeployHandlerConfig{
		Region:          os.Getenv("AWS_DEFAULT_REGION"),
		VpcID:           vpcID,
		AssignPublicIP:  "DISABLED",
		SecurityGroupID: "",
		SubnetIDs:       subnetIDs,
	})

	if err != nil {
		t.Error(err)
	}

	defer DeleteTaskRevision("figlet")
}

func TestAccCreateTaskRevision_WithSecret(t *testing.T) {
	PreTest(t)
	subnetIDs := os.Getenv("subnet_ids")
	vpcID := VpcFromSubnet(subnetIDs)

	_, err := CreateTaskRevision(requests.CreateFunctionRequest{
		Service: "hellogoworld",
		Image:   "ewilde/hellogoworld:latest",
		Secrets: []string{"db-password"},
	}, &types.DeployHandlerConfig{
		Region:          os.Getenv("AWS_DEFAULT_REGION"),
		VpcID:           vpcID,
		AssignPublicIP:  "DISABLED",
		SecurityGroupID: "",
		SubnetIDs:       subnetIDs,
	})

	if err != nil {
		t.Error(err)
	}

	defer DeleteTaskRevision("figlet")
}
