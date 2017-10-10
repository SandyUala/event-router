package main

import (
	"github.com/astronomerio/asds/api"
	"github.com/astronomerio/asds/config"
	"github.com/astronomerio/asds/marathon"
	"github.com/astronomerio/asds/mesos"
	"github.com/astronomerio/asds/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	proResourceAvailableCPU = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pro_available_cpu",
		Help: "The amount of available CPU in Pro",
	}, []string{"node"})

	proResourceAvailableMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pro_available_mem",
		Help: "The amount of available Memory in Pro",
	}, []string{"node"})

	proResourceAvailableDisk = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pro_available_disk",
		Help: "The amount of available Disk in Pro",
	}, []string{"node"})

	proResourceTotalCPU = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pro_total_cpu",
		Help: "The total amount of CPU in Pro",
	})

	proResourceTotalMem = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pro_total_mem",
		Help: "The total amount of Memory in Pro",
	})

	proResourceTotalDisk = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "pro_total_disk",
		Help: "The total amount of Disk in Pro",
	})

	proDeployCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pro_deployment_count",
		Help: "Number of pro deployments",
	}, []string{"orgid"})

	proProvisionCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pro_provision_count",
		Help: "Number of pro provisions",
	})

	proDeprovisionCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pro_deprovision_count",
		Help: "Number of pro deprovisions",
	})
)

func main() {
	// Get our configs
	config, err := config.GetConfigs()
	if err != nil {
		panic(err)
	}

	// Setup our logging
	if config.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	log := logrus.New()
	if config.LogDebug {
		log.SetLevel(logrus.DebugLevel)
	}
	logger := log.WithFields(logrus.Fields{"package": "main", "function": "main"})
	logger.Info("Starting provisioner")

	// Register Prometheus Collectors
	prometheus.MustRegister(proResourceAvailableCPU,
		proResourceAvailableDisk,
		proResourceAvailableMem,
		proResourceTotalCPU,
		proResourceTotalDisk,
		proResourceTotalMem,
		proDeployCount,
		proProvisionCount,
		proDeprovisionCount)

	pi := &pkg.PrometheusInstrumentation{
		ProResourceAvailableCPU:  *proResourceAvailableCPU,
		ProResourceAvailableMem:  *proResourceAvailableMem,
		ProResourceAvailableDisk: *proResourceAvailableDisk,
		ProResourceTotalCPU:      proResourceTotalCPU,
		ProResourceTotalDisk:     proResourceTotalDisk,
		ProResourceTotalMem:      proResourceTotalMem,
		ProDeploymentCount:       *proDeployCount,
		ProProvisionCount:        proProvisionCount,
		ProDeprovisionCount:      proDeprovisionCount,
	}

	marathonConfig, err := marathon.GetMarathonConfig()
	if err != nil {
		log.Panic(err)
	}
	marathonClient, err := marathon.GetMarathonClient(marathonConfig.MarathonURL)
	if err != nil {
		log.Panic(err)
	}
	marathonProvisioningClient, err := marathon.NewProvisioningClient(log, marathonClient)
	if err != nil {
		log.Panic(err)
	}

	mesosConfig, err := mesos.GetMesosConfig()
	if err != nil {
		log.Panic(err)
	}
	mesosClient := mesos.NewClient(mesosConfig, log, pi)

	handler := api.NewHandler(log, marathonProvisioningClient, mesosClient, pi)

	client := api.NewClient(log, config.LogDebug, handler)
	if string(config.Port[0]) != ":" {
		config.Port = ":" + config.Port
	}

	client.Serve(config.Port)
}
