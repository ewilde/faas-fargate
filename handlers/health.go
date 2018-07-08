package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// MakeHealthHandler returns 200/OK when healthy
func MakeHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		log.Info("Health check request")
		w.WriteHeader(http.StatusOK)
	}
}
