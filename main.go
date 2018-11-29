package main

import (
	"os"
	"time"

	ecsutil "github.com/ewilde/faas-fargate/aws"
	"github.com/ewilde/faas-fargate/handlers"
	"github.com/ewilde/faas-fargate/types"
	"github.com/ewilde/faas-fargate/version"
	"github.com/openfaas/faas-provider"
	bootTypes "github.com/openfaas/faas-provider/types"
	log "github.com/sirupsen/logrus"
)

func main() {
	functionNamespace := "default" // TODO: this isn't used at the moment

	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	initLogging()

	readConfig := types.ReadConfig{}
	osEnv := types.OsEnv{}
	cfg := readConfig.Read(osEnv)

	log.Infof("faas-fargate version:%s. Last commit message: %s, commit SHA: %s'", version.BuildVersion(), version.GitCommitMessage, version.GitCommitSHA)
	log.Infof("HTTP Read Timeout: %s", cfg.ReadTimeout)
	log.Infof("HTTP Write Timeout: %s", cfg.WriteTimeout)
	log.Infof("Function Readiness Probe Enabled: %v", cfg.EnableFunctionReadinessProbe)
	log.Infof("Function namespace: %v - WARNING not used at the moment", functionNamespace)

	deployConfig := &types.DeployHandlerConfig{
		AssignPublicIP:  cfg.AssignPublicIP,
		SecurityGroupID: cfg.SecurityGroupID,
		SubnetIDs:       cfg.SubnetIDs,
		Region:          cfg.DefaultAWSRegion,
		VpcID:           ecsutil.VpcFromSubnet(cfg.SubnetIDs),
	}

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.MakeProxy(cfg.ReadTimeout),
		DeleteHandler:  handlers.MakeDeleteHandler(deployConfig),
		DeployHandler:  handlers.MakeDeployHandler(deployConfig),
		FunctionReader: handlers.MakeFunctionReader(),
		ReplicaReader:  handlers.MakeReplicaReader(),
		ReplicaUpdater: handlers.MakeReplicaUpdater(),
		UpdateHandler:  handlers.MakeUpdateHandler(deployConfig),
		Health:         handlers.MakeHealthHandler(),
		InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommitSHA),
	}

	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 8,
		WriteTimeout: time.Second * 8,
		TCPPort:      &cfg.Port,
		EnableHealth: true,
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
