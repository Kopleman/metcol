package metricscollector

import (
	"errors"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockHTTP struct {
	postCallCount int
}

func (m *mockHTTP) Post(_, _ string, _ []byte) ([]byte, error) {
	m.postCallCount++
	return []byte("{}"), nil
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	mockCfg := config.Config{
		EndPoint:       nil,
		ReportInterval: 0,
		PollInterval:   0,
	}

	type fields struct {
		cfg    *config.Config
		client HTTPClient
		logger log.Logger
	}

	tests := []struct {
		fields    fields
		name      string
		numOfRuns int
		wantErr   bool
	}{
		{
			name:      "run 1",
			fields:    fields{cfg: &mockCfg, client: &mockHTTP{}, logger: log.MockLogger{}},
			numOfRuns: 1,
			wantErr:   false,
		},
		{
			name:      "run 2",
			fields:    fields{cfg: &mockCfg, client: &mockHTTP{}, logger: log.MockLogger{}},
			numOfRuns: 2,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.fields.cfg, tt.fields.logger, tt.fields.client)

			state := mc.GetState()
			assert.Equal(t, len(state), 2)
			assert.Equal(t, state["PollCount"].value, "0")

			for range tt.numOfRuns {
				err := mc.CollectAllMetrics()
				if (err != nil) != tt.wantErr {
					t.Errorf("CollectMetrics() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			afterCallState := mc.GetState()
			assert.Equal(t, 32, len(afterCallState))
			assert.Equal(t, afterCallState["PollCount"].value, strconv.Itoa(tt.numOfRuns))
		})
	}
}

func TestMetricsCollector_SendMetrics(t *testing.T) {
	mockClient := &mockHTTP{}

	mockCfg := config.Config{
		EndPoint:       nil,
		ReportInterval: 0,
		PollInterval:   0,
	}

	type fields struct {
		cfg    *config.Config
		client HTTPClient
		logger log.Logger
	}

	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:    "send metrics to endpoint",
			fields:  fields{cfg: &mockCfg, client: mockClient, logger: log.MockLogger{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.fields.cfg, tt.fields.logger, tt.fields.client)
			err := mc.CollectAllMetrics()
			if err != nil {
				t.Errorf("unwanted CollectMetrics() error = %v", err)
				return
			}

			sentErr := mc.SendMetrics()
			if (sentErr != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", sentErr, tt.wantErr)
				return
			}

			assert.Equal(t, 1, mockClient.postCallCount)
		})
	}
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Post(url, contentType string, body []byte) ([]byte, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).([]byte), args.Error(1)
}

func TestHandler(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		mc := NewMetricsCollector(
			&config.Config{
				PollInterval:   1,
				ReportInterval: 1,
			},
			log.MockLogger{},
			new(MockHTTPClient),
		)

		sig := make(chan os.Signal, 1)
		done := make(chan error)

		go func() {
			done <- mc.Handler(sig)
		}()

		time.Sleep(1500 * time.Millisecond)
		sig <- os.Interrupt

		err := <-done
		require.NoError(t, err)
	})
}

func TestIncreasePollCounter(t *testing.T) {
	mc := NewMetricsCollector(nil, nil, nil)

	require.Equal(t, "0", mc.currentMetricState[pollCountMetricName].value)

	for i := 1; i <= 5; i++ {
		require.NoError(t, mc.increasePollCounter())
		assert.Equal(t, strconv.Itoa(i), mc.currentMetricState[pollCountMetricName].value)
	}
}

func TestAssignNewRandomValue(t *testing.T) {
	rand.Seed(0)
	mc := NewMetricsCollector(nil, nil, nil)

	initial := mc.currentMetricState[randomValueMetricName].value
	mc.assignNewRandomValue()

	assert.NotEqual(t, initial, mc.currentMetricState[randomValueMetricName].value)
}

func TestSendMetricsViaWorkers(t *testing.T) {
	t.Run("worker error propagation", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		mockClient.On("Post", mock.Anything, mock.Anything, mock.Anything).
			Return([]byte{}, errors.New("worker error"))

		mc := NewMetricsCollector(&config.Config{RateLimit: 3}, nil, mockClient)
		mc.currentMetricState["test"] = MetricItem{value: "123", metricType: common.GaugeMetricType}

		err := mc.sendMetricsViaWorkers()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "worker error")
	})
}

func TestConcurrentAccess(t *testing.T) {
	mc := NewMetricsCollector(&config.Config{}, nil, nil)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				require.NoError(t, mc.CollectAllMetrics())
			}
		}()
	}
	wg.Wait()

	finalCount, err := strconv.Atoi(mc.currentMetricState[pollCountMetricName].value)
	require.NoError(t, err)
	assert.Equal(t, 1000, finalCount)
}

func TestConvertMetricItemToDto(t *testing.T) {
	tests := []struct {
		name        string
		item        MetricItem
		expectedErr string
	}{
		{
			name: "valid gauge",
			item: MetricItem{
				value:      "123.45",
				metricType: common.GaugeMetricType,
			},
		},
		{
			name: "invalid gauge",
			item: MetricItem{
				value:      "invalid",
				metricType: common.GaugeMetricType,
			},
			expectedErr: "unable to parse value",
		},
		{
			name: "valid counter",
			item: MetricItem{
				value:      "42",
				metricType: common.CounterMetricType,
			},
		},
		{
			name: "invalid counter",
			item: MetricItem{
				value:      "invalid",
				metricType: common.CounterMetricType,
			},
			expectedErr: "unable to parse value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(nil, nil, nil)
			_, err := mc.convertMetricItemToDto("test", tt.item)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
