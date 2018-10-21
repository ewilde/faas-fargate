package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func Test_ServiceNameForDisplay(t *testing.T) {
	serviceName := ServiceNameForDisplay(aws.String("openfaas-figlet"))
	if serviceName != "figlet" {
		t.Errorf("Expected figlet, actual %s", serviceName)
		t.Fail()
	}
}
