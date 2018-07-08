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

	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	awsutil "github.com/ewilde/faas-ecs/aws"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
	"fmt"
		)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(functionNamespace string, timeout time.Duration, ecsClient *ecs.ECS, ec2Client *ec2.EC2) http.HandlerFunc {
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

			hostName, err := lookupHostName(ecsClient, ec2Client, service)
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

func lookupHostName(ecsClient *ecs.ECS, ec2Client *ec2.EC2, functionName string) (string, error) {
	serviceName := awsutil.ServiceNameFromFunctionName(functionName)
	serviceArn, err := awsutil.FindECSServiceArn(ecsClient, serviceName)
	if err != nil {
		return "", err
	}

	// TODO: cache this as it's expensive to lookup for every invoke
	services, err := ecsClient.DescribeServices(&ecs.DescribeServicesInput{Cluster: awsutil.ClusterID(), Services: []*string{serviceArn}})
	if err != nil {
		return "", err
	}

	service := services.Services[0]
	if len(service.ServiceRegistries) == 0 {
		return hostNameFromENI(ecsClient, ec2Client, serviceName)
	}

	return fmt.Sprintf("%s.openfaas.local", functionName), nil

}

func hostNameFromENI(ecsClient *ecs.ECS, ec2Client *ec2.EC2, serviceName string) (string, error) {
	// TODO: cache this as it's expensive to lookup for every invoke

	listTasks, err := ecsClient.ListTasks(
		&ecs.ListTasksInput{
			Cluster:     awsutil.ClusterID(),
			ServiceName: aws.String(serviceName),
		})
	if err != nil {
		return "", err
	}

	tasks, err := ecsClient.DescribeTasks(
		&ecs.DescribeTasksInput{
			Cluster: awsutil.ClusterID(),
			Tasks:   listTasks.TaskArns,
		})
	if err != nil {
		return "", err
	}

	details := tasks.Tasks[0].Attachments[0].Details
	networkId, ok := awsutil.KeyValuePairGetValue("networkInterfaceId", details)
	if !ok {
		return "", errors.New("could not find a running task with an attached network interface")
	}

	network, err := ec2Client.DescribeNetworkInterfaces(
		&ec2.DescribeNetworkInterfacesInput{NetworkInterfaceIds: []*string{networkId}})
	if err != nil {
		return "", err
	}

	return aws.StringValue(network.NetworkInterfaces[0].Association.PublicDnsName), nil
}
