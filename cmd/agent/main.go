// Package main for run agent.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kopleman/metcol/internal/agent/config"
	metricscollector "github.com/Kopleman/metcol/internal/agent/metrics-collector"
	httpclient "github.com/Kopleman/metcol/internal/common/http-client"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	utils.PrintVersion(buildVersion, buildDate, buildCommit)

	logger := log.New(
		log.WithAppVersion("local"),
	)

	logger.Info("Starting metric collector agent")
	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger log.Logger) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	agentConfig, err := config.ParseAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to parse the agent's config: %w", err)
	}

	httpClient := httpclient.NewHTTPClient(agentConfig, logger)
	collector := metricscollector.NewMetricsCollector(agentConfig, logger, httpClient)
	if initErr := collector.Init(); initErr != nil {
		return fmt.Errorf("failed to initialize the collector: %w", initErr)
	}

	if err = collector.Handler(sig); err != nil {
		return fmt.Errorf("metrics collector error: %w", err)
	}

	return nil
}
