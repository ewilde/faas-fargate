package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	awsutil "github.com/ewilde/faas-fargate/aws"
	"github.com/ewilde/faas-fargate/types"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// MakeUpdateHandler update specified function
func MakeUpdateHandler(
	functionNamespace string,
	ecsClient *ecs.ECS,
	ec2Client *ec2.EC2,
	discovery *servicediscovery.ServiceDiscovery,
	config *types.DeployHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		taskDefinition, err := awsutil.CreateTaskRevision(ecsClient, request)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		service, err := awsutil.UpdateOrCreateECSService(ecsClient, ec2Client, discovery, taskDefinition.TaskDefinition, request, config)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		log.Infof("Updated service - %s %s.", request.Service, aws.StringValue(service.ServiceArn))
		w.WriteHeader(http.StatusAccepted)
	}
}
