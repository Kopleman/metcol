// Package main for run server.
package main

import (
	_ "net/http/pprof"

	"context"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/server"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srvConfig, err := config.ParseServerConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config for server: %w", err)
	}

	srv := server.NewServer(logger, srvConfig)
	if startErr := srv.Start(ctx); startErr != nil {
		return fmt.Errorf("start server error: %w", startErr)
	}

	return nil
}
