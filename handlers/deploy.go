// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	awsutil "github.com/ewilde/faas-ecs/aws"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/ewilde/faas-ecs/types"
	log "github.com/sirupsen/logrus"
	"fmt"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

// initialReplicasCount how many replicas to start of creating for a function
const initialReplicasCount = 1

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(
	functionNamespace string,
	ecsClient *ecs.ECS,
	ec2Client *ec2.EC2,
	discovery *servicediscovery.ServiceDiscovery,
	config *types.DeployHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		log.Infof("Deployment request for namespace %s", functionNamespace)

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			log.Errorln("Error during unmarshal of create function request. ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Infof("Deployment request for function %s", request.Service)

		taskDefinition, err := awsutil.CreateTaskRevision(ecsClient, request)
		if err != nil {
			log.Errorln(fmt.Sprintf("Error creating task revision for %s", request.Service), err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		log.Infof("Created task definition for %s", request.Service)

		service, err := awsutil.UpdateOrCreateECSService(ecsClient, ec2Client, discovery, taskDefinition.TaskDefinition, request, config)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		log.Infof("Created service %s arn: %s", request.Service, aws.StringValue(service.ServiceArn))
	}
}
