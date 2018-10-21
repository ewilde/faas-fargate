// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	awsutil "github.com/ewilde/faas-fargate/aws"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(functionNamespace string, client *ecs.ECS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Read service list for namespace %s", functionNamespace)

		functions, err := getServiceList(functionNamespace, client)
		if err != nil {
			log.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}

func getServiceList(functionNamespace string, client *ecs.ECS) ([]requests.Function, error) {
	var functions []requests.Function

	services, err := client.ListServices(
		&ecs.ListServicesInput{
			Cluster: awsutil.ClusterID(),
		})
	if err != nil {
		return nil, err
	}

	// TODO: Can only describe a maximum of 10 services in one shot
	// https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_DescribeServices.html
	// update this to deal with more that 10 services
	var serviceNames []*string
	for _, item := range services.ServiceArns {
		if !awsutil.IsFaasService(item) {
			continue
		}

		serviceNames = append(serviceNames, awsutil.ServiceNameFromArn(item))
	}

	if len(serviceNames) == 0 {
		return functions, nil
	}

	details, err := client.DescribeServices(&ecs.DescribeServicesInput{Services: serviceNames, Cluster: awsutil.ClusterID()})
	if err != nil {
		return nil, err
	}

	for _, item := range details.Services {
		task, err := client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{TaskDefinition: item.TaskDefinition})
		if err != nil {
			return nil, err
		}

		function := requests.Function{
			Name:              awsutil.ServiceNameForDisplay(item.ServiceName),
			Replicas:          uint64(*item.RunningCount),
			Image:             aws.StringValue(task.TaskDefinition.ContainerDefinitions[0].Image),
			AvailableReplicas: uint64(*item.DesiredCount), // TODO find out what this property relates to
			InvocationCount:   0,
			Labels:            nil,
		}

		functions = append(functions, function)
	}

	return functions, nil
}
