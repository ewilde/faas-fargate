package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-provider/types"
	log "github.com/sirupsen/logrus"
)

const (
	//OrchestrationIdentifier identifier string for provider orchestration
	OrchestrationIdentifier = "fargate"
	//ProviderName name of the provider
	ProviderName = "faas-fargate"
)

//MakeInfoHandler creates handler for /system/info endpoint
func MakeInfoHandler(version, sha string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		log.Info("Info request")

		infoRequest := types.InfoRequest{
			Orchestration: OrchestrationIdentifier,
			Provider:      ProviderName,
			Version: types.ProviderVersion{
				Release: version,
				SHA:     sha,
			},
		}

		jsonOut, marshalErr := json.Marshal(infoRequest)
		if marshalErr != nil {
			log.Error("Error during unmarshal of info request ", marshalErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonOut)
	}
}
