// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	awsutil "github.com/ewilde/faas-fargate/aws"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// MakeDeleteHandler delete a function
func MakeDeleteHandler(functionNamespace string, client *ecs.ECS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Delete service list for namespace %s", functionNamespace)

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			log.Errorln(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(request.FunctionName) == 0 {
			log.Errorln("Can not delete a function, request function name is empty")
			w.WriteHeader(http.StatusBadRequest)
		}

		serviceArn, err := awsutil.FindECSServiceArn(client, request.FunctionName)
		if err != nil {
			log.Errorln(fmt.Sprintf("Could not find service matching %s.", request.FunctionName), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if serviceArn == nil {
			log.Errorf("Can not delete a function, no function found matching %s", request.FunctionName)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		services, err := client.DescribeServices(&ecs.DescribeServicesInput{Cluster: awsutil.ClusterID(), Services: []*string{serviceArn}})
		if err != nil {
			log.Errorln(fmt.Sprintf("Could not describe service %s.", aws.StringValue(serviceArn)), err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		if *services.Services[0].DesiredCount > 0 {
			client.UpdateService(&ecs.UpdateServiceInput{
				Cluster:      awsutil.ClusterID(),
				Service:      serviceArn,
				DesiredCount: aws.Int64(0)})
		}

		result, err := client.DeleteService(&ecs.DeleteServiceInput{Cluster: awsutil.ClusterID(), Service: serviceArn})
		if err != nil {
			log.Errorln(fmt.Sprintf("Error deleting service %s arn: %s.", request.FunctionName, aws.StringValue(serviceArn)), err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		log.Debugf("Deleting function %s result: %s", request.FunctionName, result.String())
	}
}
