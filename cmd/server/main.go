package main

import (
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
)

func main() {
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Info("Starting server")
	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger log.Logger) error {
	srvConfig, err := config.ParseServerConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config for server: %w", err)
	}

	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)

	routes := routers.BuildServerRoutes(logger, metricsService)
	if listenAndServeErr := http.ListenAndServe(srvConfig.NetAddr.String(), routes); listenAndServeErr != nil {
		return fmt.Errorf("failed to setup server: %w", listenAndServeErr)
	}

	return nil
}
