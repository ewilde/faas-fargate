// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"net/http"

	awsutil "github.com/ewilde/faas-fargate/aws"
	log "github.com/sirupsen/logrus"
)

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		functions, err := awsutil.GetServiceList()

		if err != nil {
			log.Errorf("Error reading functions.\n%v", err)
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
