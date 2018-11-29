// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"fmt"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(timeout time.Duration) http.HandlerFunc {
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			// MaxIdleConns:          1,
			// DisableKeepAlives:     false,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodGet,
			http.MethodPost:

			vars := mux.Vars(r)
			service := vars["name"]

			stamp := strconv.FormatInt(time.Now().Unix(), 10)

			defer func(when time.Time) {
				seconds := time.Since(when).Seconds()
				log.Printf("[%s] took %f seconds\n", stamp, seconds)
			}(time.Now())

			forwardReq := requests.NewForwardRequest(r.Method, *r.URL)

			hostName, err := lookupHostName(service)
			if err != nil {
				log.Errorln(fmt.Sprintf("Error looking up host name using service %s. ", service), err)
				writeError(err, service, w)
				return
			}

			url := forwardReq.ToURL(hostName, watchdogPort)

			request, _ := http.NewRequest(r.Method, url, r.Body)

			copyHeaders(&request.Header, &r.Header)

			defer request.Body.Close()

			response, err := proxyClient.Do(request)
			if err != nil {
				writeError(err, service, w)
				return
			}

			clientHeader := w.Header()
			copyHeaders(&clientHeader, &response.Header)

			writeHead(service, http.StatusOK, w)
			io.Copy(w, response.Body)
		}
	}
}

func writeError(err error, service string, w http.ResponseWriter) {
	log.Println(err.Error())
	writeHead(service, http.StatusInternalServerError, w)
	buf := bytes.NewBufferString("Can't reach service: " + service) // TODO: print error in body?
	w.Write(buf.Bytes())
}

func writeHead(service string, code int, w http.ResponseWriter) {
	w.WriteHeader(code)
}

func copyHeaders(destination *http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(*destination)[k] = vClone
	}
}

func lookupHostName(functionName string) (string, error) {
	return fmt.Sprintf("%s.openfaas.local", functionName), nil
}
