package controllers_test

import (
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
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/Kopleman/metcol/internal/server/sterrors"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockMetricsService struct {
	GetValueAsStringFn func(ctx context.Context, metricType common.MetricType, name string) (string, error)
	GetMetricAsDTOFn   func(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error)
}

func (m *MockMetricsService) GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error) {
	return m.GetValueAsStringFn(ctx, metricType, name)
}

func (m *MockMetricsService) GetMetricAsDTO(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error) {
	return m.GetMetricAsDTOFn(ctx, metricType, name)
}

func TestGetValueController_GetValue(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		mockSetup      func(*MockMetricsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "invalid metric type",
			url:  "/value/invalid_type/metric1",
			mockSetup: func(ms *MockMetricsService) {
				ms.GetValueAsStringFn = func(ctx context.Context, metricType common.MetricType, name string) (string, error) {
					return "", errors.New("not used")
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unknown metric type\n",
		},
		{
			name: "empty metric name",
			url:  "/value/gauge/",
			mockSetup: func(ms *MockMetricsService) {
				ms.GetValueAsStringFn = func(ctx context.Context, metricType common.MetricType, name string) (string, error) {
					return "", nil
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name: "metric not found",
			url:  "/value/gauge/nonexistent",
			mockSetup: func(ms *MockMetricsService) {
				ms.GetValueAsStringFn = func(ctx context.Context, metricType common.MetricType, name string) (string, error) {
					return "", sterrors.ErrNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			name: "internal server error",
			url:  "/value/counter/metric1",
			mockSetup: func(ms *MockMetricsService) {
				ms.GetValueAsStringFn = func(ctx context.Context, metricType common.MetricType, name string) (string, error) {
					return "", errors.New("some error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "something went wrong\n",
		},
		{
			name: "success case",
			url:  "/value/gauge/metric1",
			mockSetup: func(ms *MockMetricsService) {
				ms.GetValueAsStringFn = func(ctx context.Context, metricType common.MetricType, name string) (string, error) {
					return "123.45", nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			// Инициализация роутера chi
			router := chi.NewRouter()
			ms := &MockMetricsService{}
			if tt.mockSetup != nil {
				tt.mockSetup(ms)
			}
			ctrl := controllers.NewGetValueController(&log.MockLogger{}, ms)

			router.Get("/value/{metricType}/{metricName}", ctrl.GetValue())
			router.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestGetValueController_GetValueAsDTO(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockMetricsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid json",
			requestBody:    `{invalid}`,
			mockSetup:      func(ms *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unable to parse dto\n",
		},
		{
			name:        "metric not found",
			requestBody: `{"id":"metric1", "type":"gauge"}`,
			mockSetup: func(ms *MockMetricsService) {
				ms.GetMetricAsDTOFn = func(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error) {
					return nil, sterrors.ErrNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			name:        "success case",
			requestBody: `{"id":"metric1", "type":"gauge"}`,
			mockSetup: func(ms *MockMetricsService) {
				ms.GetMetricAsDTOFn = func(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error) {
					return &dto.MetricDTO{
						ID:    "metric1",
						MType: "gauge",
						Value: testutils.Pointer(123.45),
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"metric1","type":"gauge","value":123.45}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/value", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			ms := &MockMetricsService{}
			if tt.mockSetup != nil {
				tt.mockSetup(ms)
			}
			ctrl := controllers.NewGetValueController(&log.MockLogger{}, ms)

			handler := ctrl.GetValueAsDTO()
			handler(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response dto.MetricDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				var expectedResponse dto.MetricDTO
				err = json.Unmarshal([]byte(tt.expectedBody), &expectedResponse)
				require.NoError(t, err)

				assert.Equal(t, expectedResponse, response)
			} else {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
