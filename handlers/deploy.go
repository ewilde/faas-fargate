// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsutil "github.com/ewilde/faas-fargate/aws"
	"github.com/ewilde/faas-fargate/types"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

// initialReplicasCount how many replicas to start of creating for a function
const initialReplicasCount = 1

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(
	config *types.DeployHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			log.Errorln("Error during unmarshal of create function request. ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Infof("Deployment request for function %s", request.Service)

		taskDefinition, err := awsutil.CreateTaskRevision(request, config)
		if err != nil {
			log.Errorln(fmt.Sprintf("Error creating task revision for %s", request.Service), err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		log.Infof("Created task definition for %s", request.Service)

		service, err := awsutil.UpdateOrCreateECSService(taskDefinition.TaskDefinition, request, config)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		log.Infof("Created service %s arn: %s", request.Service, aws.StringValue(service.ServiceArn))
	}
}
