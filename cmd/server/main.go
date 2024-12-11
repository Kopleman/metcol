package main

import (
	"fmt"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/gofiber/fiber/v3"
)

func main() {
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer logger.Sync()

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

	app := fiber.New()
	routers.BuildAppRoutes(logger, app, metricsService)

	addr := srvConfig.NetAddr.String()
	logger.Infof("Server started on %s", addr)
	if listenAndServeErr := app.Listen(addr); listenAndServeErr != nil {
		return fmt.Errorf("failed to setup server: %w", listenAndServeErr)
	}

	return nil
}
