package routers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/server/metrics"
	mock "github.com/Kopleman/metcol/internal/server/routers/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

type mockBodyDecryptor struct{}

func (p *mockBodyDecryptor) DecryptBody(body io.Reader) (io.Reader, error) {
	return body, nil
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)
	if body != http.NoBody {
		req.Header.Set(common.ContentType, "application/json")
	}

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
	ctrl := gomock.NewController(t)
	mockPgx := mock.NewMockPgxPool(ctrl)
	mockBD := &mockBodyDecryptor{}

	storeService := memstore.NewStore(make(map[string]*dto.MetricDTO))
	metricsService := metrics.NewMetrics(storeService, log.MockLogger{})
	routes := BuildServerRoutes(&config.Config{}, &log.MockLogger{}, metricsService, mockPgx, mockBD)

	ts := httptest.NewServer(routes)
	defer ts.Close()

	var testTable = []struct {
		method  string
		url     string
		body    io.Reader
		want    string
		status  int
		hasJSON bool
	}{
		{"POST", "/update/gauge/testGauge/100", http.NoBody, "", http.StatusOK, false},
		{"POST", "/update/counter/testCounter/100", http.NoBody, "", http.StatusOK, false},
		{"POST", "/update/gauge/badGauge/nope", http.NoBody, "can not parse input value\n", http.StatusBadRequest, false},
		{"GET", "/update/counter/testCounter/100", http.NoBody, "Method Not Allowed\n", http.StatusMethodNotAllowed, false},
		{"GET", "/value/gauge/testGauge", http.NoBody, "100", http.StatusOK, false},
		{
			"GET",
			"/value/gauge/testUnknown94",
			http.NoBody,
			"failed to read metric 'testunknown94': not found\n",
			http.StatusNotFound,
			false,
		},
		{"GET", "/", http.NoBody, "testcounter:100\ntestgauge:100\n", http.StatusOK, false},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "gauge", "value": 1.2}`),
			`{"id":"foo","value":1.2,"type":"gauge"}`,
			http.StatusOK,
			true,
		},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "counter", "delta": 100}`),
			`{"id":"foo","delta":100,"type":"counter"}`,
			http.StatusOK,
			true,
		},
		{
			"POST",
			"/updates/",
			strings.NewReader(`[{"id": "baz", "type": "counter", "delta": 100}, {"id": "bar", "type": "gauge", "value": 1.2}]`),
			`[{"id":"baz","delta":100,"type":"counter"}, {"id": "bar", "type": "gauge", "value": 1.2}]`,
			http.StatusOK,
			true,
		},
		{
			"POST",
			"/update",
			strings.NewReader(`{"id": "foo", "type": "counter", "value": "nope"}`),
			"unable to parse dto\n",
			http.StatusBadRequest,
			false,
		},
		{
			"POST",
			"/value",
			strings.NewReader(`{"id": "foo", "type": "counter"}`),
			`{"id":"foo","delta":100,"type":"counter"}`,
			http.StatusOK,
			true,
		},
		{
			"POST",
			"/value",
			strings.NewReader(`{"id": "foo", "type": "gauge"}`),
			`{"id":"foo","value":1.2,"type":"gauge"}`,
			http.StatusOK,
			true,
		},
	}
	for _, v := range testTable {
		gotStatusCode, gotResponse := testRequest(t, ts, v.method, v.url, v.body)
		assert.Equal(t, v.status, gotStatusCode)

		if !v.hasJSON {
			assert.Equal(t, v.want, gotResponse)
			continue
		}

		assert.JSONEq(t, v.want, gotResponse)
	}
}
