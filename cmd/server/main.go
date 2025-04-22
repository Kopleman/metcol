// Package main for run server.
package main

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"context"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/server"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	utils.PrintVersion(buildVersion, buildDate, buildCommit)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := log.New(
		log.WithAppVersion("local"),
		log.WithLogLevel(log.INFO),
	)
	defer logger.Sync() //nolint:all // its safe

	onErrChan := make(chan error)
	defer close(onErrChan)
	srv := run(ctx, logger, onErrChan)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// Wait system context done or onError
		for {
			select {
			case err := <-onErrChan:
				if err != nil {
					logger.Infof("Starting graceful server shut down due to error: %s", err.Error())
					return err
				}
			case <-ctx.Done():
				logger.Info("Starting graceful server shut down")
				return nil
			}
		}
	})
	g.Go(func() error {
		<-gCtx.Done()
		srv.Shutdown()
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf("server shut down unexpectedly due to: %w", err)
	}
}

func run(ctx context.Context, logger log.Logger, onErrorChan chan<- error) *server.Server {
	srvConfig, err := config.ParseServerConfig()
	if err != nil {
		logger.Fatalf("failed to parse config for server: %w", err)
	}

	srv := server.NewServer(logger, srvConfig)

	// Start server
	go func(ctx context.Context) {
		runTimeError := make(chan error, 1)
		defer close(runTimeError)

		if serverStartError := srv.Start(ctx, runTimeError); serverStartError != nil {
			onErrorChan <- fmt.Errorf("failed to start server: %w", serverStartError)
		}
		serverRunTimeError := <-runTimeError
		if serverRunTimeError != nil {
			onErrorChan <- fmt.Errorf("runtime error: %w", serverRunTimeError)
		}
	}(ctx)

	return srv
}
