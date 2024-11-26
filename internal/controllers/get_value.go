package controllers

import (
	"errors"
	"fmt"
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/store"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetValue(metricsService metrics.IMetrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		metricTypeStringAsString := strings.ToLower(chi.URLParam(req, "metricType"))
		metricType, err := metrics.ParseMetricType(metricTypeStringAsString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricName := strings.ToLower(chi.URLParam(req, "metricName"))
		if len(metricName) == 0 {
			http.Error(w, "empty metric name", http.StatusNotFound)
			return
		}

		fmt.Println(fmt.Printf("getValue called with metricType='%s', metricName='%s' at %s", metricType, metricName, time.Now().UTC()))

		value, err := metricsService.GetValueAsString(metricType, metricName)

		if err != nil && !errors.Is(err, store.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.WriteString(w, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
