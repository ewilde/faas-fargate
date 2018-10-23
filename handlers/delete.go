// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

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

		err = awsutil.DeleteECSService(client, request.FunctionName)
		if err != nil {
			log.Errorf("Can not delete function %s. %v", request.FunctionName, err)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
