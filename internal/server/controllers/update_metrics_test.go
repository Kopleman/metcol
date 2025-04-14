package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SetMetric(
	ctx context.Context,
	metricType common.MetricType,
	name string,
	value string,
) error {
	args := m.Called(ctx, metricType, name, value)
	return args.Error(0) //nolint:wrapcheck // mocked-err
}

func (m *MockMetricsService) SetMetricByDto(ctx context.Context, metricDto *dto.MetricDTO) error {
	args := m.Called(ctx, metricDto)
	return args.Error(0) //nolint:wrapcheck // mocked-err
}

func (m *MockMetricsService) SetMetrics(ctx context.Context, metrics []*dto.MetricDTO) error {
	args := m.Called(ctx, metrics)
	return args.Error(0) //nolint:wrapcheck // mocked-err
}

type MockBodyDecryptor struct {
	mock.Mock
}

func (m *MockBodyDecryptor) DecryptBody(body io.Reader) (io.Reader, error) {
	args := m.Called(body)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.Reader), args.Error(1)
}

func TestUpdateOrSet(t *testing.T) {
	mockService := new(MockMetricsService)
	mockLogger := new(log.MockLogger)
	mockDecrypter := new(MockBodyDecryptor)
	ctrl := NewUpdateMetricsController(mockLogger, mockService, mockDecrypter)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		mockError      error
	}{
		{
			name:           "success gauge",
			url:            "/update/gauge/test_metric/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid metric type",
			url:            "/update/invalid/test_metric/123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty metric name",
			url:            "/update/counter//123",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty metric value",
			url:            "/update/counter/test_metric/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "value parse error",
			url:            "/update/counter/test_metric/invalid",
			expectedStatus: http.StatusBadRequest,
			mockError:      metrics.ErrValueParse,
		},
		{
			name:           "internal server error",
			url:            "/update/counter/test_metric/123",
			expectedStatus: http.StatusInternalServerError,
			mockError:      errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, tt.url, http.NoBody)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			parts := strings.Split(strings.TrimPrefix(tt.url, "/"), "/")
			rctx.URLParams = chi.RouteParams{}
			rctx.URLParams.Add("metricType", parts[1])
			rctx.URLParams.Add("metricName", parts[2])
			rctx.URLParams.Add("metricValue", parts[3])
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.mockError != nil || tt.expectedStatus == http.StatusOK {
				mockService.On("SetMetric", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.mockError).Once()
			}

			ctrl.UpdateOrSet()(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateOrSetViaDTO(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockDecryptErr error
		mockServiceErr error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           `{"id":"test","type":"gauge","value":123.45}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "decrypt error",
			mockDecryptErr: errors.New("decrypt error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `invalid`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			body:           `{"id":"test","type":"gauge","value":123.45}`,
			mockServiceErr: errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptor := new(MockBodyDecryptor)
			decryptor.On("DecryptBody", mock.Anything).Return(strings.NewReader(tt.body), tt.mockDecryptErr)

			service := new(MockMetricsService)
			service.On("SetMetricByDto", mock.Anything, mock.Anything).Return(tt.mockServiceErr)

			ctrl := NewUpdateMetricsController(
				log.MockLogger{},
				service,
				decryptor,
			)

			req := httptest.NewRequest("POST", "/update", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			ctrl.UpdateOrSetViaDTO()(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}

	t.Run("decryption error handling", func(t *testing.T) {
		decryptor := new(MockBodyDecryptor)
		decryptor.On("DecryptBody", mock.Anything).
			Return(nil, errors.New("decryption failed"))

		ctrl := NewUpdateMetricsController(
			log.MockLogger{},
			new(MockMetricsService),
			decryptor,
		)

		req := httptest.NewRequest("POST", "/update", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		ctrl.UpdateOrSetViaDTO()(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateMetrics(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockDecryptErr error
		mockServiceErr error
		expectedStatus int
	}{
		{
			name:           "success batch",
			body:           `[{"id":"test1","type":"gauge","value":1}]`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "decrypt error",
			body:           `invalid`,
			mockDecryptErr: errors.New("decrypt error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json after decrypt",
			body:           `invalid`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			body:           `[{"id":"test","type":"gauge","value":123.45}]`,
			mockServiceErr: metrics.ErrValueParse,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal server error",
			body:           `[{"id":"test","type":"gauge","value":123.45}]`,
			mockServiceErr: errors.New("unexpected error"),
			expectedStatus: http.StatusBadRequest, // Изменилось с 500 на 400 согласно новому коду
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptor := new(MockBodyDecryptor)
			decryptor.On("DecryptBody", mock.Anything).
				Return(strings.NewReader(tt.body), tt.mockDecryptErr)

			service := new(MockMetricsService)
			service.On("SetMetrics", mock.Anything, mock.Anything).Return(tt.mockServiceErr)

			ctrl := NewUpdateMetricsController(
				log.MockLogger{},
				service,
				decryptor,
			)

			req := httptest.NewRequest("POST", "/updates", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			ctrl.UpdateMetrics()(w, req)

			resp := w.Result()
			defer resp.Body.Close() //nolint:all // tests

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var result []dto.MetricDTO
				err := json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestParseUpdateBody(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      string
		mockDecryptErr error
		expectError    bool
	}{
		{
			name:      "valid batch",
			inputBody: `[{"id":"cpu","type":"gauge","value":42.5}]`,
		},
		{
			name:      "valid single",
			inputBody: `{"id":"cpu","type":"gauge","value":42.5}`,
		},
		{
			name:           "decrypt error",
			inputBody:      `invalid`,
			mockDecryptErr: errors.New("decrypt error"),
			expectError:    true,
		},
		{
			name:        "invalid json",
			inputBody:   `invalid`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptor := new(MockBodyDecryptor)
			decryptor.On("DecryptBody", mock.Anything).
				Return(strings.NewReader(tt.inputBody), tt.mockDecryptErr)

			ctrl := NewUpdateMetricsController(
				log.MockLogger{},
				new(MockMetricsService),
				decryptor,
			)

			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.inputBody))
			_, err := ctrl.parseUpdateBody(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
