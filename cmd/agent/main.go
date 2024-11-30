package main

import (
	"fmt"
	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/agent/metrics-collector"
	"github.com/Kopleman/metcol/internal/common/http-client"
	"github.com/Kopleman/metcol/internal/common/log"
)

func main() {
	logger := log.New(
		log.WithAppVersion("local"),
	)

	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger log.Logger) error {
	agentConfig, err := config.ParseAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to parse the agent's config: %w", err)
	}

	httpClient := httpclient.NewHTTPClient(agentConfig, logger)
	collector := metricscollector.NewMetricsCollector(agentConfig, logger, httpClient)

	collector.Run()

	return nil
}
