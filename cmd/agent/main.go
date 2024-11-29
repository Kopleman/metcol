package main

import (
	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/agent/metrics-collector"
	"github.com/Kopleman/metcol/internal/common/http-client"
	"github.com/Kopleman/metcol/internal/common/log"
)

func main() {
	logger := log.New(
		log.WithAppVersion("local"),
	)

	agentConfig, err := config.ParseAgentConfig()
	if err != nil {
		logger.Fatal(err)
	}

	httpClient := httpclient.NewHTTPClient(agentConfig)
	collector := metricscollector.NewMetricsCollector(agentConfig, logger, httpClient)

	collector.Run()
}
