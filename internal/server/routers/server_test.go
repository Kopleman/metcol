package routers

import (
	"errors"
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		err = errors.Join(err, resp.Body.Close())
		require.NoError(t, err)
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}

func TestRouters_Server(t *testing.T) {
	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)
	routes := BuildServerRoutes(metricsService)

	ts := httptest.NewServer(routes)
	defer ts.Close()

	var testTable = []struct {
		method string
		url    string
		want   string
		status int
	}{
		{"POST", "/update/gauge/testGauge/100", "", http.StatusOK},
		{"POST", "/update/counter/testCounter/100", "", http.StatusOK},
		{"POST", "/update/gauge/badGauge/nope", "can not parse input value\n", http.StatusBadRequest},
		{"GET", "/update/counter/testCounter/100", "Only POST requests are allowed!\n", http.StatusMethodNotAllowed},
		{"GET", "/value/gauge/testGauge", "100", http.StatusOK},
		{"GET", "/value/gauge/testUnknown94", "not found\n", http.StatusNotFound},
		{"GET", "/", "testgauge:100\ntestcounter:100\n", http.StatusOK},
	}
	for _, v := range testTable {
		gotStatusCode, gotResponse := testRequest(t, ts, v.method, v.url)
		assert.Equal(t, v.status, gotStatusCode)
		assert.Equal(t, v.want, gotResponse)
	}
}
