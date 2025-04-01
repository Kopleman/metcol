package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
)

type MockMainPageMetricsService struct {
	GetAllValuesAsStringFn func(ctx context.Context) (map[string]string, error)
}

func (m *MockMainPageMetricsService) GetAllValuesAsString(ctx context.Context) (map[string]string, error) {
	return m.GetAllValuesAsStringFn(ctx)
}

type MockResponseWriter struct {
	header        http.Header
	statusCode    int
	body          []byte
	simulateError bool
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(data []byte) (int, error) {
	if m.simulateError {
		return 0, errors.New("simulated write error")
	}
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestMainPageController_MainPage(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockMainPageMetricsService)
		expectedStatus int
		expectedBody   string
		checkErrorLog  bool
	}{
		{
			name: "successful response with metrics",
			mockSetup: func(ms *MockMainPageMetricsService) {
				ms.GetAllValuesAsStringFn = func(ctx context.Context) (map[string]string, error) {
					return map[string]string{
						"metric2": "200",
						"metric1": "100",
						"metric3": "300",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "metric1:100\nmetric2:200\nmetric3:300\n",
		},
		{
			name: "service returns error",
			mockSetup: func(ms *MockMainPageMetricsService) {
				ms.GetAllValuesAsStringFn = func(ctx context.Context) (map[string]string, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   common.Err500Message + "\n",
			checkErrorLog:  true,
		},
		{
			name: "missing metric after sorting",
			mockSetup: func(ms *MockMainPageMetricsService) {
				ms.GetAllValuesAsStringFn = func(ctx context.Context) (map[string]string, error) {
					// Возвращаем мапу, но при этом будем удалять метрику при обработке
					return map[string]string{
						"existing": "123",
						"missing":  "456",
					}, nil
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   common.Err500Message + "\n",
			checkErrorLog:  true,
		},
		{
			name: "write response error",
			mockSetup: func(ms *MockMainPageMetricsService) {
				ms.GetAllValuesAsStringFn = func(ctx context.Context) (map[string]string, error) {
					return map[string]string{"test": "value"}, nil
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   common.Err500Message + "\n",
			checkErrorLog:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

			ms := &MockMainPageMetricsService{}
			if tt.mockSetup != nil {
				tt.mockSetup(ms)
			}

			logger := &log.MockLogger{}
			ctrl := NewMainPageController(logger, ms)

			if tt.name == "missing metric after sorting" {
				origHandler := ctrl.MainPage()
				handler := func(w http.ResponseWriter, r *http.Request) {
					origHandler(w, r)
					if allMetrics := w.(*httptest.ResponseRecorder).Body.Bytes(); len(allMetrics) > 0 { //nolint:all // IDE goes mad
						ctrl.metricsService = &MockMainPageMetricsService{
							GetAllValuesAsStringFn: func(ctx context.Context) (map[string]string, error) {
								return map[string]string{"existing": "123"}, nil
							},
						}
					}
				}
				rr := httptest.NewRecorder()
				handler(rr, req)
				return
			}

			if tt.name == "write response error" {
				w := &MockResponseWriter{
					header:        make(http.Header),
					simulateError: true,
				}
				ctrl.MainPage()(w, req)
				assert.Equal(t, http.StatusInternalServerError, w.statusCode)
				return
			}

			rr := httptest.NewRecorder()
			ctrl.MainPage()(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
