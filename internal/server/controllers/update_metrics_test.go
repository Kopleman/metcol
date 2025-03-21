package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SetMetric(ctx context.Context, metricType common.MetricType, name string, value string) error {
	args := m.Called(ctx, metricType, name, value)
	return args.Error(0)
}

func (m *MockMetricsService) SetMetricByDto(ctx context.Context, metricDto *dto.MetricDTO) error {
	args := m.Called(ctx, metricDto)
	return args.Error(0)
}

func (m *MockMetricsService) SetMetrics(ctx context.Context, metrics []*dto.MetricDTO) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func TestUpdateOrSet(t *testing.T) {
	mockService := new(MockMetricsService)
	mockLogger := new(log.MockLogger)
	ctrl := NewUpdateMetricsController(mockLogger, mockService)

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
			r := httptest.NewRequest("POST", tt.url, nil)
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
	mockService := new(MockMetricsService)
	mockLogger := new(log.MockLogger)
	ctrl := NewUpdateMetricsController(mockLogger, mockService)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		mockError      error
	}{
		{
			name: "success",
			payload: dto.MetricDTO{
				ID:    "test_metric",
				MType: "gauge",
				Value: testutils.Pointer(123.45),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid payload",
			payload:        "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "value parse error",
			payload: dto.MetricDTO{
				ID:    "test_metric",
				MType: "gauge",
				Value: testutils.Pointer(123.45),
			},
			expectedStatus: http.StatusBadRequest,
			mockError:      metrics.ErrValueParse,
		},
		{
			name: "internal error",
			payload: dto.MetricDTO{
				ID:    "test_metric",
				MType: "gauge",
				Value: testutils.Pointer(123.45),
			},
			expectedStatus: http.StatusInternalServerError,
			mockError:      errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.payload); err != nil {
				require.NoError(t, err)
			}

			r := httptest.NewRequest("POST", "/update", &body)
			w := httptest.NewRecorder()

			if tt.mockError != nil || tt.expectedStatus == http.StatusOK {
				mockService.On("SetMetricByDto", mock.Anything, mock.AnythingOfType("*dto.MetricDTO")).Return(tt.mockError).Once()
			}

			ctrl.UpdateOrSetViaDTO()(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var response dto.MetricDTO
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				assert.Equal(t, tt.payload, response)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	mockService := new(MockMetricsService)
	mockLogger := new(log.MockLogger)
	ctrl := NewUpdateMetricsController(mockLogger, mockService)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		mockError      error
	}{
		{
			name: "success batch",
			payload: []*dto.MetricDTO{
				{ID: "metric1", MType: "gauge", Value: testutils.Pointer(1.23)},
				{ID: "metric2", MType: "counter", Delta: testutils.Pointer(int64(42))},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success single",
			payload: &dto.MetricDTO{
				ID:    "test_metric",
				MType: "gauge",
				Value: testutils.Pointer(123.45),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid payload",
			payload:        "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			payload: []*dto.MetricDTO{
				{ID: "metric1", MType: "gauge", Value: testutils.Pointer(1.23)},
			},
			expectedStatus: http.StatusBadRequest,
			mockError:      metrics.ErrValueParse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.payload); err != nil {
				require.NoError(t, err)
			}

			r := httptest.NewRequest("POST", "/updates", &body)
			w := httptest.NewRecorder()

			if tt.mockError != nil || tt.expectedStatus == http.StatusOK {
				mockService.On("SetMetrics", mock.Anything, mock.Anything).Return(tt.mockError).Once()
			}

			ctrl.UpdateMetrics()(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var response interface{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestParseUpdateBody(t *testing.T) {
	ctrl := UpdateMetricsController{}

	tests := []struct {
		name        string
		payload     interface{}
		expectedLen int
		expectError bool
	}{
		{
			name: "valid batch",
			payload: []*dto.MetricDTO{
				{
					Value: testutils.Pointer(123.45),
					ID:    "metric1",
					MType: "gauge",
				},
				{
					ID:    "metric2",
					MType: "counter",
					Delta: testutils.Pointer(int64(42)),
				},
			},
			expectedLen: 2,
		},
		{
			name: "valid single",
			payload: &dto.MetricDTO{
				Value: testutils.Pointer(123.45),
				ID:    "metric1",
				MType: "gauge",
			},
			expectedLen: 1,
		},
		{
			name:        "invalid payload",
			payload:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			require.NoError(t, json.NewEncoder(&body).Encode(tt.payload))

			r := httptest.NewRequest("POST", "/", &body)
			result, err := ctrl.parseUpdateBody(r)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(result))
			}
		})
	}
}
