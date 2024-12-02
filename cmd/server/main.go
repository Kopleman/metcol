package main

import (
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
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
	srvConfig, err := config.ParseServerConfig()
	if err != nil {
		return err
	}

	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)
	routes := routers.BuildServerRoutes(logger, metricsService)

	if listenAndServeErr := http.ListenAndServe(srvConfig.NetAddr.String(), routes); listenAndServeErr != nil {
		return fmt.Errorf("failed to setup server: %v", listenAndServeErr)
	}

	return nil
}
