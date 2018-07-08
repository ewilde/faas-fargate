package main

import (
	"time"

	"github.com/openfaas/faas-provider"

	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	ecsutil "github.com/ewilde/faas-ecs/aws"
	"github.com/ewilde/faas-ecs/handlers"
	"github.com/ewilde/faas-ecs/types"
	"github.com/ewilde/faas-ecs/version"
	bootTypes "github.com/openfaas/faas-provider/types"
	log "github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	functionNamespace := "default"

	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	initLogging()

	readConfig := types.ReadConfig{}
	osEnv := types.OsEnv{}
	cfg := readConfig.Read(osEnv)

	log.Infof("Faas-ecs version:%s. Last commit message: %s, commit SHA: %s'", version.BuildVersion(), version.GitCommitMessage, version.GitCommitSHA)
	log.Infof("HTTP Read Timeout: %s", cfg.ReadTimeout)
	log.Infof("HTTP Write Timeout: %s", cfg.WriteTimeout)
	log.Infof("Function Readiness Probe Enabled: %v", cfg.EnableFunctionReadinessProbe)

	session := session.Must(session.NewSession())
	ecsClient := ecs.New(session, aws.NewConfig().WithLogLevel(awsLogLevel()))
	ec2Client := ec2.New(session, aws.NewConfig().WithLogLevel(awsLogLevel()))
	discovery := servicediscovery.New(session, aws.NewConfig().WithLogLevel(awsLogLevel()))

	deployConfig := &types.DeployHandlerConfig{
		AssignPublicIP: cfg.AssignPublicIP,
		SecurityGroupId: cfg.SecurityGroupId,
		SubnetIDs: cfg.SubnetIDs,
		VpcID: ecsutil.VpcFromSubnet(ec2Client, cfg.SubnetIDs),
	}


	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.MakeProxy(functionNamespace, cfg.ReadTimeout, ecsClient, ec2Client),
		DeleteHandler:  handlers.MakeDeleteHandler(functionNamespace, ecsClient),
		DeployHandler:  handlers.MakeDeployHandler(functionNamespace, ecsClient, ec2Client, discovery, deployConfig),
		FunctionReader: handlers.MakeFunctionReader(functionNamespace, ecsClient),
		ReplicaReader:  handlers.MakeReplicaReader(functionNamespace, ecsClient),
		ReplicaUpdater: handlers.MakeReplicaUpdater(functionNamespace, ecsClient),
		UpdateHandler:  handlers.MakeUpdateHandler(functionNamespace, ecsClient, ec2Client, discovery, deployConfig),
		Health:         handlers.MakeHealthHandler(),
		InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommitSHA),
	}

	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 8,
		WriteTimeout: time.Second * 8,
		TCPPort:      &cfg.Port,
	}

	log.Infof("Listening on port %d", cfg.Port)
	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}

func initLogging() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to debug
	if !ok {
		lvl = "info"
	}

	// parse string, this is built-in feature of logrus
	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)

	log.Debugf("Logging level set to %v", ll.String())
}

func awsLogLevel() aws.LogLevelType {
	lvl := os.Getenv("LOG_LEVEL")

	if lvl == "trace" {
		return aws.LogDebugWithRequestErrors
	}

	return aws.LogOff
}
