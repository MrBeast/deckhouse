/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"d8.io/upmeter/pkg/agent"
	"d8.io/upmeter/pkg/kubernetes"
	"d8.io/upmeter/pkg/server"
)

func parseServerArgs(cmd *kingpin.CmdClause, config *server.Config) {
	// Serve
	cmd.Flag("listen-host", "Upmeter server host.").
		Envar("UPMETER_SERVICE_HOST").
		Default("localhost").
		StringVar(&config.ListenHost)

	cmd.Flag("listen-port", "Upmeter server port.").
		Envar("UPMETER_SERVICE_PORT").
		Default("8091").
		StringVar(&config.ListenPort)

	// Database
	cmd.Flag("db-path", "SQLite file path.").
		Envar("UPMETER_DB_PATH").
		Default("upmeter.db").
		StringVar(&config.DatabasePath)

	// Origins count
	cmd.Flag("origins", "The expected number of origins, used for exporting episodes as metrics when they are fulfilled by this number of agents.").
		Required().
		Envar("UPMETER_ORIGINS").
		IntVar(&config.OriginsCount)
}

func parseAgentArgs(cmd *kingpin.CmdClause, config *agent.Config) {
	// Sender
	cmd.Flag("service-host", "Upmeter server host.").
		Envar("UPMETER_SERVICE_HOST").
		Default("localhost").
		StringVar(&config.ClientConfig.Host)

	cmd.Flag("service-port", "Upmeter server port.").
		Envar("UPMETER_SERVICE_PORT").
		Default("8091").
		StringVar(&config.ClientConfig.Port)

	cmd.Flag("ca-path", "CA path").
		Envar("UPMETER_CA_PATH").
		Default("").
		StringVar(&config.ClientConfig.CAPath)

	cmd.Flag("tls", "Should we use TLS").
		Envar("UPMETER_TLS").
		Default("false").
		BoolVar(&config.ClientConfig.TLS)

	cmd.Flag("period", "The period of episodes sending to server, and at the same the client timeout.").
		Envar("UPMETER_PERIOD").
		Default("1s").
		DurationVar(&config.Period)

	// Database
	cmd.Flag("db-path", "SQLite file path.").
		Envar("UPMETER_DB_PATH").
		Default("upmeter.db").
		StringVar(&config.DatabasePath)
}

func parseKubeArgs(cmd *kingpin.CmdClause, config *kubernetes.Config) {
	cmd.Flag("kube-context", "The name of the kubeconfig context to use. Can be set with $KUBE_CONTEXT.").
		Envar("KUBE_CONTEXT").
		Default("").
		StringVar(&config.Context)

	cmd.Flag("kube-config", "Path to the kubeconfig file. Can be set with $KUBE_CONFIG.").
		Envar("KUBE_CONFIG").
		Default("").
		StringVar(&config.Config)

	cmd.Flag("kube-server", "The address and port of the Kubernetes API server. Can be set with $KUBE_SERVER.").
		Envar("KUBE_SERVER").
		Default("").
		StringVar(&config.Server)

	// Rate limit settings for kube client
	cmd.Flag("kube-client-qps", "QPS for a rate limiter of a kubernetes client. Can be set with $KUBE_CLIENT_QPS.").
		Envar("KUBE_CLIENT_QPS").
		Default("5"). // DefaultQPS from k8s.io/client-go/rest/config.go
		Float32Var(&config.ClientQps)

	cmd.Flag("kube-client-burst", "Burst for a rate limiter of a kubernetes client. Can be set with $KUBE_CLIENT_BURST.").
		Envar("KUBE_CLIENT_BURST").
		Default("10"). // DefaultBurst from k8s.io/client-go/rest/config.go
		IntVar(&config.ClientBurst)

	cmd.Flag("scheduler-probe-image", "Image for control plane scheduler probe").
		Envar("UPMETER_SCHEDULER_PROBE_IMAGE").
		Default(kubernetes.DefaultAlpineImage).
		StringVar(&config.SchedulerProbeImage.Name)

	cmd.Flag("scheduler-probe-pull-secrets", "Image pull secrets names for control plane scheduler image").
		Envar("UPMETER_SCHEDULER_PROBE_IMAGE_PULL_SECRETS").
		Default("").
		StringsVar(&config.SchedulerProbeImage.PullSecrets)

	cmd.Flag("scheduler-probe-node", "Node to schedule the pod to").
		Envar("UPMETER_SCHEDULER_PROBE_NODE").
		Default("").
		StringVar(&config.SchedulerProbeNode)

	cmd.Flag("ccm-namespace", "Cloud Controller Manager namespace").
		Envar("UPMETER_CLOUD_CONTROLLER_MANAGER_NAMESPACE").
		Default("").
		StringVar(&config.CloudControllerManagerNamespace)

	cmd.Flag("cluster-domain", "Cluster domain").
		Envar("UPMETER_CLUSTER_DOMAIN").
		Default("cluster.local").
		StringVar(&config.ClusterDomain)
}

type loggerConfig struct {
	Level  string
	NoTime bool
	Type   string
}

// SetupLoggingSettings init global flags for logging
func parseLoggerArgs(cmd *kingpin.CmdClause, config *loggerConfig) {
	cmd.Flag("log-level", "Logging level: debug, info, error. Default is info. Can be set with $LOG_LEVEL.").
		Envar("LOG_LEVEL").
		Default("info").
		StringVar(&config.Level)
	cmd.Flag("log-type", "Logging formatter type: json, text or color. Default is text. Can be set with $LOG_TYPE.").
		Envar("LOG_TYPE").
		Default("text").
		StringVar(&config.Type)
	cmd.Flag("log-no-time", "Disable timestamp logging if flag is present. Useful when output is redirected to logging system that already adds timestamps. Can be set with $LOG_NO_TIME.").
		Envar("LOG_NO_TIME").
		BoolVar(&config.NoTime)
}

// setupLogger sets logging output
func setupLogger(logger *log.Logger, config *loggerConfig) {
	switch config.Type {
	case "json":
		logger.SetFormatter(&log.JSONFormatter{DisableTimestamp: config.NoTime})
	case "text":
		logger.SetFormatter(&log.TextFormatter{DisableTimestamp: config.NoTime, DisableColors: true})
	case "color":
		logger.SetFormatter(&log.TextFormatter{DisableTimestamp: config.NoTime, ForceColors: true})
	default:
		logger.SetFormatter(&log.JSONFormatter{DisableTimestamp: config.NoTime})
	}

	switch strings.ToLower(config.Level) {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
}
