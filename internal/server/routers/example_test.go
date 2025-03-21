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

func (p *noopPgxPool) Ping(err context.Context) error {
	return nil
}

func setupRouter(baseMemStore map[string]*dto.MetricDTO) *chi.Mux {
	storeService := memstore.NewStore(baseMemStore)
	metricsService := metrics.NewMetrics(storeService, log.MockLogger{})
	mockPgx := &noopPgxPool{}
	routes := BuildServerRoutes(&config.Config{}, &log.MockLogger{}, metricsService, mockPgx)
	return routes
}

func ExampleBuildServerRoutes_updateMetric() {
	router := setupRouter(make(map[string]*dto.MetricDTO))

	ts := httptest.NewServer(router)
	defer ts.Close()

	// Update metric via URL params
	paramReq, err := http.NewRequest("POST", ts.URL+"/update/gauge/testerrmetric/123.45", nil)
	if err != nil {
		fmt.Println(err)
	}
	paramsResp, err := http.DefaultClient.Do(paramReq)
	if err != nil {
		fmt.Println(err)
	}
	defer paramsResp.Body.Close()

	// Update metric via JSON
	jsonBody := []byte(`{
		"id": "testerrcounter",
		"type": "counter",
		"delta": 10
	}`)
	jsonReq, err := http.NewRequest("POST", ts.URL+"/update", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println(err)
	}
	jsonReq.Header.Set("Content-Type", "application/json")
	jsonResp, err := http.DefaultClient.Do(jsonReq)
	if err != nil {
		fmt.Println(err)
	}
	body, err := io.ReadAll(jsonResp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonResp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// {"delta":10,"id":"testerrcounter","type":"counter"}
}

func ExampleBuildServerRoutes_getMetric() {
	mockStore := map[string]*dto.MetricDTO{
		"testerrmetric-gauge": {
			ID:    "testerrmetric",
			MType: "gauge",
			Delta: nil,
			Value: testutils.Pointer(1.1),
		},
		"testerrmetric-counter": {
			ID:    "testerrmetric",
			MType: "counter",
			Delta: testutils.Pointer(int64(100)),
			Value: nil,
		},
	}
	router := setupRouter(mockStore)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Fetch metric via URL params
	paramReq, err := http.NewRequest("GET", ts.URL+"/value/counter/testerrmetric", nil)
	if err != nil {
		fmt.Println(err)
	}
	paramResp, err := http.DefaultClient.Do(paramReq)
	if err != nil {
		fmt.Println(err)
	}
	body, err := io.ReadAll(paramResp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer paramResp.Body.Close()
	fmt.Println(string(body))

	// Fetch metric via JSON
	jsonBody := []byte(`{
		"id": "testerrmetric",
		"type": "gauge"
	}`)
	jsonReq, err := http.NewRequest("POST", ts.URL+"/value", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println(err)
	}
	jsonReq.Header.Set("Content-Type", "application/json")
	jsonResp, err := http.DefaultClient.Do(jsonReq)
	if err != nil {
		fmt.Println(err)
	}
	body, err = io.ReadAll(jsonResp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonResp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// 100
	// {"value":1.1,"id":"testerrmetric","type":"gauge"}
}

func ExampleBuildServerRoutes_batchUpdate() {
	router := setupRouter(make(map[string]*dto.MetricDTO))
	ts := httptest.NewServer(router)
	defer ts.Close()

	jsonBody := []byte(`[{
		"id": "testerrcounter",
		"type": "counter",
		"delta": 5
	}, {
		"id": "testerrgauge",
		"type": "gauge",
		"value": 3.14
	}]`)

	req, err := http.NewRequest("POST", ts.URL+"/updates", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	fmt.Println(string(body))

	// Output:
	// [{"delta":5,"id":"testerrcounter","type":"counter"},{"value":3.14,"id":"testerrgauge","type":"gauge"}]
}
