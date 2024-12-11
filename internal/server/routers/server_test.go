package routers

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, app *fiber.App, method,
	path string) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, path, http.NoBody)
	require.NoError(t, err)

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
		want   string
		status int
	}{
		{"POST", "/update/gauge/testGauge/100", "OK", http.StatusOK},
		{"POST", "/update/counter/testCounter/100", "OK", http.StatusOK},
		{"POST", "/update/gauge/badGauge/nope", "can not parse input value", http.StatusBadRequest},
		{"GET", "/update/counter/testCounter/100", "Method Not Allowed", http.StatusMethodNotAllowed},
		{"GET", "/value/gauge/testGauge", "100", http.StatusOK},
		{
			"GET",
			"/value/gauge/testUnknown94",
			"failed to read metric 'testunknown94-gauge': not found",
			http.StatusNotFound,
		},
		{"GET", "/", "testcounter:100\ntestgauge:100\n", http.StatusOK},
	}
	for _, v := range testTable {
		gotStatusCode, gotResponse := testRequest(t, app, v.method, v.url)
		assert.Equal(t, v.status, gotStatusCode)
		assert.Equal(t, v.want, gotResponse)
	}
}
