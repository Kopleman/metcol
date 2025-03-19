//nolint:all // lots of stuff
package routers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/go-chi/chi/v5"
)

type noopPgxPool struct{}

func (p *noopPgxPool) Ping(_ context.Context) error {
	return nil
}

func setupRouter(baseMemStore map[string]*dto.MetricDTO) *chi.Mux {
	storeService := memstore.NewStore(baseMemStore)
	metricsService := metrics.NewMetrics(storeService, log.MockLogger{})
	mockPgx := &noopPgxPool{}
	routes := BuildServerRoutes(&config.Config{}, &log.MockLogger{}, metricsService, mockPgx)
	return routes
}

func ExampleUpdateMetric() {
	router := setupRouter(make(map[string]*dto.MetricDTO))

	ts := httptest.NewServer(router)
	defer ts.Close()

	// Update metric via URL params
	paramReq, _ := http.NewRequest("POST", ts.URL+"/update/gauge/test_metric/123.45", nil)
	paramsResp, _ := http.DefaultClient.Do(paramReq)
	defer paramsResp.Body.Close()

	// Update metric via JSON
	jsonBody := []byte(`{
		"id": "test_counter",
		"type": "counter",
		"delta": 10
	}`)
	jsonReq, _ := http.NewRequest("POST", ts.URL+"/update", bytes.NewBuffer(jsonBody))
	jsonReq.Header.Set("Content-Type", "application/json")
	jsonResp, _ := http.DefaultClient.Do(jsonReq)
	body, _ := io.ReadAll(jsonResp.Body)
	defer jsonResp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// {"delta":10,"id":"test_counter","type":"counter"}
}

func ExampleGetMetric() {
	mockStore := map[string]*dto.MetricDTO{
		"test_metric-gauge": {
			ID:    "test_metric",
			MType: "gauge",
			Delta: nil,
			Value: testutils.Pointer(1.1),
		},
		"test_metric-counter": {
			ID:    "test_metric",
			MType: "counter",
			Delta: testutils.Pointer(int64(100)),
			Value: nil,
		},
	}
	router := setupRouter(mockStore)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Fetch metric via URL params
	paramReq, _ := http.NewRequest("GET", ts.URL+"/value/counter/test_metric", nil)
	paramResp, _ := http.DefaultClient.Do(paramReq)
	body, _ := io.ReadAll(paramResp.Body)
	defer paramResp.Body.Close()
	fmt.Println(string(body))

	// Fetch metric via JSON
	jsonBody := []byte(`{
		"id": "test_metric",
		"type": "gauge"
	}`)
	jsonReq, _ := http.NewRequest("POST", ts.URL+"/value", bytes.NewBuffer(jsonBody))
	jsonReq.Header.Set("Content-Type", "application/json")
	jsonResp, _ := http.DefaultClient.Do(jsonReq)
	body, _ = io.ReadAll(jsonResp.Body)
	defer jsonResp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// 100
	// {"value":1.1,"id":"test_metric","type":"gauge"}
}

func ExampleBatchUpdate() {
	router := setupRouter(make(map[string]*dto.MetricDTO))
	ts := httptest.NewServer(router)
	defer ts.Close()

	jsonBody := []byte(`[{
		"id": "test_counter",
		"type": "counter",
		"delta": 5
	}, {
		"id": "test_gauge",
		"type": "gauge",
		"value": 3.14
	}]`)

	req, _ := http.NewRequest("POST", ts.URL+"/updates", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// [{"delta":5,"id":"test_counter","type":"counter"},{"value":3.14,"id":"test_gauge","type":"gauge"}]
}

func ExamplePing() {
	mockStore := map[string]*dto.MetricDTO{
		"test_metric_gauge-gauge": {
			ID:    "test_metric_gauge",
			MType: "gauge",
			Delta: nil,
			Value: testutils.Pointer(1.1),
		},
		"test_metric_counter-counter": {
			ID:    "test_metric_counter",
			MType: "counter",
			Delta: testutils.Pointer(int64(100)),
			Value: nil,
		},
	}
	router := setupRouter(mockStore)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/")
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// test_metric_counter:100
	// test_metric_gauge:1.1
}
