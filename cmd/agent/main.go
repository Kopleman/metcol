package main

import (
	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/agent/metrics-collector"
	"github.com/Kopleman/metcol/internal/common/http-client"
	"github.com/Kopleman/metcol/internal/common/log"
)

func main() {
	agentConfig := config.ParseAgentConfig()

	logger := log.New(
		log.WithAppVersion("local"),
	)

	httpClient := httpclient.NewHTTPClient(agentConfig)
	collector := metricscollector.NewMetricsCollector(agentConfig, logger, httpClient)

	// Init logger

	collector.Run()
}
