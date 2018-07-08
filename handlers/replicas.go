// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-netes/types"
	"github.com/openfaas/faas/gateway/requests"

	"github.com/aws/aws-sdk-go/aws"
	awsutil "github.com/ewilde/faas-ecs/aws"
	log "github.com/sirupsen/logrus"
)

// MakeReplicaUpdater updates desired count of replicas
func MakeReplicaUpdater(functionNamespace string, client *ecs.ECS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Update replicas")

		request := types.ScaleServiceRequest{}
		if r.Body != nil {
			defer r.Body.Close()
			bytesIn, _ := ioutil.ReadAll(r.Body)
			marshalErr := json.Unmarshal(bytesIn, &request)
			if marshalErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				msg := "Cannot parse request. Please pass valid JSON."
				w.Write([]byte(msg))
				log.Errorln(msg, marshalErr)
				return
			}
		}

		service, err := awsutil.UpdateECSServiceDesiredCount(client, request.ServiceName, int(request.Replicas))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		log.Infof("Updated service replica count for %s with arn %s", request.ServiceName, aws.StringValue(service.ServiceArn))

	}
}

// MakeReplicaReader reads the amount of replicas for a deployment
func MakeReplicaReader(functionNamespace string, client *ecs.ECS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Read replicas")

		vars := mux.Vars(r)
		functionName := vars["name"]

		functions, err := getServiceList(functionNamespace, client)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var found *requests.Function
		for _, function := range functions {
			if function.Name == functionName {
				found = &function
				break
			}
		}

		if found == nil {
			w.WriteHeader(404)
			return
		}

		functionBytes, _ := json.Marshal(found)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}
