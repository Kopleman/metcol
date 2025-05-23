// Package server bootstrap server.
package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/profiler"
	bodydecryptor "github.com/Kopleman/metcol/internal/server/body_decryptor"
	"github.com/Kopleman/metcol/internal/server/config"
	filestorage "github.com/Kopleman/metcol/internal/server/file_storage"
	"github.com/Kopleman/metcol/internal/server/grpc"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/pgxstore"
	"github.com/Kopleman/metcol/internal/server/postgres"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
)

// Server instance of server.
type Server struct {
	logger        log.Logger
	config        *config.Config
	db            *postgres.PostgreSQL
	store         store.Store
	fs            *filestorage.FileStorage
	metricService *metrics.Metrics
	bd            *bodydecryptor.BodyDecryptor
	grpcServer    *grpc.Server
}

// NewServer creates instance of server.
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

	s.logger.Info("no db DSN provided, using memo-store")
	storeService := memstore.NewStore(make(map[string]*dto.MetricDTO))
	s.store = storeService
	s.metricService = metrics.NewMetrics(s.store, s.logger)
	s.fs = filestorage.NewFileStorage(s.config, s.logger, s.metricService)
	if err := s.fs.Init(ctx); err != nil {
		return fmt.Errorf("failed to init filestorage: %w", err)
	}
	return nil
}

// Start starts new server.
func (s *Server) Start(ctx context.Context, runTimeError chan<- error) error {
	if err := s.prepareStore(ctx); err != nil {
		return fmt.Errorf("failed to prepare store: %w", err)
	}

	bd := bodydecryptor.NewBodyDecryptor(s.logger)
	if err := bd.LoadPrivateKey(s.config.PrivateKeyPath); err != nil {
		return fmt.Errorf("failed to init bodyDecryptor: %w", err)
	}
	s.bd = bd

	if s.fs != nil {
		go func(ctx context.Context) {
			err := s.fs.RunBackupJob(ctx)
			if err != nil {
				runTimeError <- fmt.Errorf("backup job error: %w", err)
			}
		}(ctx)
	}

	go func() {
		routes := routers.BuildServerRoutes(s.config, s.logger, s.metricService, s.db, s.bd)
		if listenAndServeErr := http.ListenAndServe(s.config.NetAddr.String(), routes); listenAndServeErr != nil {
			runTimeError <- fmt.Errorf("internal server error: %w", listenAndServeErr)
		}
	}()

	if s.config.GRPCAddr.String() != "" {
		grpcMetricsService := grpc.NewMetricsService(s.logger, s.metricService)
		s.grpcServer = grpc.NewServer(s.logger, grpcMetricsService, s.config.TrustedSubnet, s.config.Key)

		go func() {
			if err := s.grpcServer.Start(s.config.GRPCAddr.String()); err != nil {
				runTimeError <- fmt.Errorf("internal server error: %w", err)
			}
		}()
	}

	s.logger.Infof("Server started on: %s", s.config.NetAddr.Port)

	go func() {
		s.logger.Info("Starting collect profiles")
		if err := profiler.Collect(profiler.Config{
			CPUProfilePath: s.config.ProfilerCPUFilePath,
			MemProfilePath: s.config.ProfilerMemFilePath,
			CollectTime:    s.config.ProfilerCollectTime,
		}); err != nil {
			runTimeError <- fmt.Errorf("failed to collect profiles: %w", err)
		}
		s.logger.Info("Finished collect profiles")
	}()

	return nil
}

// Shutdown called on shutdown.
func (s *Server) Shutdown() {
	if s.fs != nil {
		s.fs.Close()
		s.logger.Info("file storage closed")
	}

	if s.db != nil {
		s.db.Close()
		s.logger.Info("database closed")
	}

	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.logger.Info("grpc server stopped")
	}

	s.logger.Infof("Server shut down")
}
