package filestorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
)

type Store interface {
	Create(key string, value any) error
	GetAll() map[string]any
}

type MetricService interface {
	ExportMetrics() ([]*dto.MetricDTO, error)
	ImportMetrics(metricsToImport []*dto.MetricDTO) error
}

type FileStorage struct {
	cfg           *config.Config
	logger        log.Logger
	metricService MetricService
	file          *os.File
	encoder       *json.Encoder
	decoder       *json.Decoder
}

func (fs *FileStorage) ExportMetrics() error {
	if err := fs.file.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate file store: %w", err)
	}
	metricsAsDTO, exportError := fs.metricService.ExportMetrics()
	if exportError != nil {
		return fmt.Errorf("could not export metrics: %w", exportError)
	}
	storeErr := fs.encoder.Encode(metricsAsDTO)
	if storeErr != nil {
		return fmt.Errorf("could not store data: %w", storeErr)
	}
	return nil
}

func (fs *FileStorage) ImportMetrics() error {
	metricsData := make([]*dto.MetricDTO, 0)
	if err := fs.decoder.Decode(&metricsData); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("could not decode metrics data from file: %w", err)
	}

	if err := fs.metricService.ImportMetrics(metricsData); err != nil {
		return fmt.Errorf("could not re-store data to store: %w", err)
	}

	return nil
}

func (fs *FileStorage) Init() error {
	file, err := os.OpenFile(fs.cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666) //nolint:all // different lint behavior on perm var
	if err != nil {
		return fmt.Errorf("could not open storage file: %w", err)
	}
	fs.file = file
	fs.encoder = json.NewEncoder(file)
	fs.decoder = json.NewDecoder(file)

	if fs.cfg.Restore {
		if reStoreErr := fs.ImportMetrics(); reStoreErr != nil {
			return fmt.Errorf("could not re-store data: %w", reStoreErr)
		}
	}
	return nil
}

func (fs *FileStorage) Close() {
	if err := fs.ExportMetrics(); err != nil {
		fs.logger.Errorf("could not store data to file: %w", err)
	}
	if err := fs.file.Close(); err != nil {
		fs.logger.Errorf("could not close storage file: %w", err)
	}
}

type intervalJobsArg struct {
	storeTimer      time.Time
	storeInterval   time.Duration
	storeInProgress bool
}

func (fs *FileStorage) RunBackupJob() error {
	now := time.Now()

	storeDuration := time.Duration(fs.cfg.StoreInterval) * time.Second

	args := intervalJobsArg{
		storeTimer:      now.Add(storeDuration),
		storeInterval:   storeDuration,
		storeInProgress: false,
	}

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan bool)

	for {
		select {
		case <-ticker.C:
			go fs.doStoreInterval(&args, quit)
		case <-quit:
			ticker.Stop()
			return errors.New("backup failed")
		}
	}
}

func (fs *FileStorage) doStoreInterval(args *intervalJobsArg, quit chan bool) {
	now := time.Now()
	if !now.After(args.storeTimer) {
		return
	}
	args.storeInProgress = true

	err := fs.ExportMetrics()
	if err != nil {
		fs.logger.Errorf("could not store data to file in interval: %w", err)
		quit <- true
		return
	}
	fs.logger.Info("stored data to file in interval")

	args.storeTimer = now.Add(args.storeInterval)

	args.storeInProgress = false
}

func NewFileStorage(cfg *config.Config, logger log.Logger, service MetricService) *FileStorage {
	return &FileStorage{
		cfg:           cfg,
		logger:        logger,
		metricService: service,
	}
}
