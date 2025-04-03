package filestorage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockMetricService struct {
	mock.Mock
}

func (m *MockMetricService) ExportMetrics(ctx context.Context) ([]*dto.MetricDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*dto.MetricDTO), args.Error(1)
}

func (m *MockMetricService) ImportMetrics(ctx context.Context, metrics []*dto.MetricDTO) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func TestFileStorage_ExportMetrics(t *testing.T) {
	t.Run("successful export", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-export-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		metrics := []*dto.MetricDTO{
			{ID: "test1", MType: "gauge", Value: testutils.Pointer(123.45)},
			{ID: "test2", MType: "counter", Delta: testutils.Pointer(int64(42))},
		}

		mockService := new(MockMetricService)
		mockService.On("ExportMetrics", mock.Anything).Return(metrics, nil)

		cfg := &config.Config{FileStoragePath: tmpFile.Name()}
		fs := NewFileStorage(cfg, nil, mockService)
		err = fs.Init(context.Background())
		require.NoError(t, err)

		err = fs.ExportMetrics()
		require.NoError(t, err)

		var result []*dto.MetricDTO
		_, err = tmpFile.Seek(0, 0)
		require.NoError(t, err)
		err = json.NewDecoder(tmpFile).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, metrics, result)
	})

	t.Run("export error", func(t *testing.T) {
		mockService := new(MockMetricService)
		mockService.On("ExportMetrics", mock.Anything).Return([]*dto.MetricDTO{}, errors.New("export error"))

		cfg := &config.Config{FileStoragePath: "/dev/null"}
		fs := NewFileStorage(cfg, nil, mockService)
		err := fs.Init(context.Background())
		require.NoError(t, err)

		err = fs.ExportMetrics()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "export error")
	})
}

func TestFileStorage_ImportMetrics(t *testing.T) {
	t.Run("successful import", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-import-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		metrics := []*dto.MetricDTO{
			{ID: "test1", MType: "gauge", Value: testutils.Pointer(123.45)},
		}

		err = json.NewEncoder(tmpFile).Encode(metrics)
		require.NoError(t, err)
		tmpFile.Close()

		mockService := new(MockMetricService)
		mockService.On("ImportMetrics", mock.Anything, metrics).Return(nil)

		cfg := &config.Config{
			FileStoragePath: tmpFile.Name(),
			Restore:         true,
		}

		fs := NewFileStorage(cfg, nil, mockService)
		err = fs.Init(context.Background())
		require.NoError(t, err)

		mockService.AssertExpectations(t)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-empty-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		cfg := &config.Config{
			FileStoragePath: tmpFile.Name(),
			Restore:         true,
		}

		fs := NewFileStorage(cfg, nil, new(MockMetricService))
		err = fs.Init(context.Background())
		require.NoError(t, err)
	})
}

func TestFileStorage_Init(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		tmpFile := "/tmp/non-existent-file.json"
		defer os.Remove(tmpFile)

		cfg := &config.Config{FileStoragePath: tmpFile}
		fs := NewFileStorage(cfg, nil, new(MockMetricService))
		err := fs.Init(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, fs.file)
	})

	t.Run("restore disabled", func(t *testing.T) {
		mockService := new(MockMetricService)
		cfg := &config.Config{
			Restore:         false,
			StoreInterval:   1,
			FileStoragePath: "/dev/null",
		}

		fs := NewFileStorage(cfg, nil, mockService)
		err := fs.Init(context.Background())
		require.NoError(t, err)
		mockService.AssertNotCalled(t, "ImportMetrics")
	})
}

func TestFileStorage_Close(t *testing.T) {
	t.Run("close with export", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-close-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		mockService := new(MockMetricService)
		mockService.On("ExportMetrics", mock.Anything).Return([]*dto.MetricDTO{}, nil)

		mockLogger := new(log.MockLogger)
		cfg := &config.Config{FileStoragePath: tmpFile.Name()}

		fs := NewFileStorage(cfg, mockLogger, mockService)
		err = fs.Init(context.Background())
		require.NoError(t, err)

		fs.Close()
		mockService.AssertExpectations(t)
	})
}
