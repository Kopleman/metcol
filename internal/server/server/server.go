package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	filestorage "github.com/Kopleman/metcol/internal/server/file_storage"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/pgxstore"
	"github.com/Kopleman/metcol/internal/server/postgres"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
)

type Server struct {
	logger        log.Logger
	config        *config.Config
	db            *postgres.PostgreSQL
	store         store.Store
	fs            *filestorage.FileStorage
	metricService *metrics.Metrics
}

func NewServer(logger log.Logger, cfg *config.Config) *Server {
	s := &Server{
		logger: logger,
		config: cfg,
	}

	return s
}

func (s *Server) prepareStore(ctx context.Context) error {
	if s.config.DataBaseDSN != "" {
		if err := postgres.RunMigrations(s.config.DataBaseDSN); err != nil {
			return fmt.Errorf("failed to run migrations on store prepare: %w", err)
		}

		pg, err := postgres.NewPostgresSQL(ctx, s.logger, s.config.DataBaseDSN)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		s.db = pg
		s.store = pgxstore.NewPGXStore(s.logger, s.db)
		s.metricService = metrics.NewMetrics(s.store, s.logger)
		return nil
	}

	storeService := memstore.NewStore(make(map[string]*dto.MetricDTO))
	s.store = storeService
	s.metricService = metrics.NewMetrics(s.store, s.logger)
	s.fs = filestorage.NewFileStorage(s.config, s.logger, s.metricService)
	if err := s.fs.Init(ctx); err != nil {
		return fmt.Errorf("failed to init filestorage: %w", err)
	}
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	defer s.Shutdown()
	if err := s.prepareStore(ctx); err != nil {
		return fmt.Errorf("failed to prepare store: %w", err)
	}

	runTimeError := make(chan error, 1)
	if s.fs != nil {
		go func() {
			err := s.fs.RunBackupJob()
			if err != nil {
				runTimeError <- fmt.Errorf("backup job error: %w", err)
			}
		}()
	}

	go func() {
		routes := routers.BuildServerRoutes(s.logger, s.metricService, s.db)
		if listenAndServeErr := http.ListenAndServe(s.config.NetAddr.String(), routes); listenAndServeErr != nil {
			runTimeError <- fmt.Errorf("internal server error: %w", listenAndServeErr)
		}
	}()
	s.logger.Infof("Server started on: %s", s.config.NetAddr.Port)

	serverError := <-runTimeError
	if serverError != nil {
		return fmt.Errorf("server error: %w", serverError)
	}

	<-ctx.Done()

	return nil
}

func (s *Server) Shutdown() {
	if s.fs != nil {
		s.fs.Close()
	}

	if s.db != nil {
		s.db.Close()
	}

	s.logger.Infof("Server shut down")
}
