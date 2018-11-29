// Copyright (c) Edward Wilde 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Package types contains definitions for public types
package types

import (
	"os"
	"strconv"
	"time"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// HasEnv provides interface for os.Getenv
type HasEnv interface {
	Getenv(key string) string
}

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func parseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}

	return duration
}

func parseBoolValue(val string, fallback bool) bool {
	if len(val) > 0 {
		return val == "true"
	}
	return fallback
}

func parseString(val string, fallback string) string {
	if len(val) > 0 {
		return val
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) BootstrapConfig {
	defaultTCPPort := 8080

	cfg := BootstrapConfig{}

	cfg.EnableFunctionReadinessProbe = parseBoolValue(hasEnv.Getenv("enable_function_readiness_probe"), true)
	cfg.ReadTimeout = parseIntOrDurationValue(hasEnv.Getenv("read_timeout"), time.Second*10)
	cfg.WriteTimeout = parseIntOrDurationValue(hasEnv.Getenv("write_timeout"), time.Second*10)
	cfg.ImagePullPolicy = parseString(hasEnv.Getenv("image_pull_policy"), "Always")
	cfg.AssignPublicIP = parseString(hasEnv.Getenv("assign_public_ip"), "DISABLED")
	cfg.Port = parseIntValue(hasEnv.Getenv("port"), defaultTCPPort)
	cfg.SubnetIDs = parseString(hasEnv.Getenv("subnet_ids"), "")
	cfg.SecurityGroupID = parseString(hasEnv.Getenv("security_group_id"), "")
	cfg.DefaultAWSRegion = parseString(hasEnv.Getenv("AWS_DEFAULT_REGION"), "us-east-1")

	return cfg
}

// BootstrapConfig for the process.
type BootstrapConfig struct {
	AssignPublicIP               string
	EnableFunctionReadinessProbe bool
	ImagePullPolicy              string
	Port                         int
	ReadTimeout                  time.Duration
	SubnetIDs                    string
	SecurityGroupID              string
	WriteTimeout                 time.Duration
	DefaultAWSRegion             string
}
