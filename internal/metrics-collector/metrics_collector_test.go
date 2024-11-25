package metrics_collector

import (
	"fmt"
	htttp_client "github.com/Kopleman/metcol/internal/http-client"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type mockHttp struct {
	postCallCount int
}

func (m *mockHttp) Post(url, contentType string, body io.Reader) ([]byte, error) {
	m.postCallCount++
	return make([]byte, 0), nil
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	type fields struct {
		client htttp_client.IHttpClient
	}

	tests := []struct {
		name      string
		fields    fields
		numOfRuns int
	}{
		{
			name:      "run 1",
			fields:    fields{&mockHttp{}},
			numOfRuns: 1,
		},
		{
			name:      "run 2",
			fields:    fields{&mockHttp{}},
			numOfRuns: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.fields.client)

			state := mc.GetState()
			assert.Equal(t, len(state), 2)
			assert.Equal(t, state["PollCount"].value, "0")

			for i := 0; i < tt.numOfRuns; i++ {
				mc.CollectMetrics()
			}

			afterCallState := mc.GetState()
			assert.Equal(t, len(afterCallState), 29)
			assert.Equal(t, afterCallState["PollCount"].value, fmt.Sprintf("%v", tt.numOfRuns))
		})
	}
}

func TestMetricsCollector_SendMetrics(t *testing.T) {
	mockClient := &mockHttp{}
	type fields struct {
		client htttp_client.IHttpClient
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "send metrics to endpoint",
			fields:  fields{mockClient},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.fields.client)
			mc.CollectMetrics()

			err := mc.SendMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, mockClient.postCallCount, 29)
		})
	}
}
