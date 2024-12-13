package routers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, app *fiber.App, method,
	path string, body io.Reader) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)
	if body != http.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, -1)

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
	mockLogger := log.MockLogger{}
	app := fiber.New()
	BuildAppRoutes(mockLogger, app, metricsService)

	var testTable = []struct {
		method string
		url    string
		body   io.Reader
		want   string
		status int
	}{
		{"POST", "/update/gauge/testGauge/100", http.NoBody, "OK", http.StatusOK},
		{"POST", "/update/counter/testCounter/100", http.NoBody, "OK", http.StatusOK},
		{"POST", "/update/gauge/badGauge/nope", http.NoBody, "can not parse input value", http.StatusBadRequest},
		{"GET", "/update/counter/testCounter/100", http.NoBody, "Method Not Allowed", http.StatusMethodNotAllowed},
		{"GET", "/value/gauge/testGauge", http.NoBody, "100", http.StatusOK},
		{
			"GET",
			"/value/gauge/testUnknown94",
			http.NoBody,
			"failed to read metric 'testunknown94-gauge': not found",
			http.StatusNotFound,
		},
		{"GET", "/", http.NoBody, "testcounter:100\ntestgauge:100\n", http.StatusOK},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "gauge", "value": 1.2}`),
			`{"id":"foo","value":1.2,"type":"gauge"}`,
			http.StatusOK,
		},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "counter", "delta": 100}`),
			`{"id":"foo","delta":100,"type":"counter"}`,
			http.StatusOK,
		},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "counter", "value": "nope"}`),
			`unable to parse dto`,
			http.StatusBadRequest,
		},
		{
			"POST",
			"/value",
			strings.NewReader(`{"id": "foo", "type": "counter"}`),
			`{"id":"foo","delta":100,"type":"counter"}`,
			http.StatusOK,
		},
		{
			"POST",
			"/value",
			strings.NewReader(`{"id": "foo", "type": "gauge"}`),
			`{"id":"foo","value":1.2,"type":"gauge"}`,
			http.StatusOK,
		},
	}
	for _, v := range testTable {
		gotStatusCode, gotResponse := testRequest(t, app, v.method, v.url, v.body)
		assert.Equal(t, v.status, gotStatusCode)
		assert.Equal(t, v.want, gotResponse)
	}
}
