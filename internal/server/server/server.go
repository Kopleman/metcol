package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	filestorage "github.com/Kopleman/metcol/internal/server/file_storage"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/postgres"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
)

type Server struct {
	logger log.Logger
	config *config.Config
	db     *postgres.PostgreSQL
}

func NewServer(logger log.Logger, cfg *config.Config) *Server {
	s := &Server{
		logger: logger,
		config: cfg,
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	defer s.Shutdown()
	pg, err := postgres.NewPostgresSQL(ctx, s.logger, s.config.DataBaseDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)
	fs := filestorage.NewFileStorage(s.config, s.logger, metricsService)
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
		routes := routers.BuildServerRoutes(s.logger, metricsService, pg)
		if listenAndServeErr := http.ListenAndServe(s.config.NetAddr.String(), routes); listenAndServeErr != nil {
			runTimeError <- fmt.Errorf("internal server error: %w", listenAndServeErr)
		}
	}()
	s.logger.Infof("Server started on: %s", s.config.NetAddr.Port)

	serverError := <-runTimeError
	if serverError != nil {
		return fmt.Errorf("server error: %w", serverError)
	}

	// Wait context
	<-ctx.Done()

	return nil
}

func (s *Server) Shutdown() {
	// stop database
	if s.db != nil {
		s.db.Close()
	}

	s.logger.Infof("Server shut down")
}
