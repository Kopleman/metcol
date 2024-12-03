package metricscollector

import (
	"io"
	"strconv"
	"testing"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
)

type mockHTTP struct {
	postCallCount int
}

func (m *mockHTTP) Post(_, _ string, _ io.Reader) ([]byte, error) {
	m.postCallCount++
	return make([]byte, 0), nil
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
				err := mc.CollectMetrics()
				if (err != nil) != tt.wantErr {
					t.Errorf("CollectMetrics() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			afterCallState := mc.GetState()
			assert.Equal(t, len(afterCallState), 29)
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
			err := mc.CollectMetrics()
			if err != nil {
				t.Errorf("unwanted CollectMetrics() error = %v", err)
				return
			}

			sentErr := mc.SendMetrics()
			if (sentErr != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", sentErr, tt.wantErr)
				return
			}

			assert.Equal(t, 29, mockClient.postCallCount)
		})
	}
}
