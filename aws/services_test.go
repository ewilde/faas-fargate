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

func Test_ServiceList(t *testing.T) {
	PreTest(t)
	services, err := GetServiceList()
	if err != nil {
		t.Error(err)
	}

	if len(services) == 0 {
		t.Errorf("Expected more than 0 services")
	}
}
