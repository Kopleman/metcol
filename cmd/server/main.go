package main

import (
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	filestorage "github.com/Kopleman/metcol/internal/server/file_storage"
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
			logger.Errorf("Error syncing logger: %v", err)
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
	fs := filestorage.NewFileStorage(srvConfig, logger, metricsService)
	if err = fs.Init(); err != nil {
		return fmt.Errorf("failed to init filestorage: %w", err)
	}
	defer fs.Close()

	runTimeError := make(chan error, 1)
	go func() {
		err = fs.RunBackupJob()
		if err != nil {
			runTimeError <- fmt.Errorf("backup job error: %w", err)
		}
	}()

	go func() {
		routes := routers.BuildServerRoutes(logger, metricsService)
		if listenAndServeErr := http.ListenAndServe(srvConfig.NetAddr.String(), routes); listenAndServeErr != nil {
			runTimeError <- fmt.Errorf("internal server error: %w", listenAndServeErr)
		}
	}()
	logger.Infof("Server started on: %s", srvConfig.NetAddr.Port)

	serverError := <-runTimeError
	if serverError != nil {
		return fmt.Errorf("server error: %w", serverError)
	}

	return nil
}
