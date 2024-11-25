package controllers

import (
	"errors"
	"fmt"
	"github.com/Kopleman/metcol/internal/metrics"
	"net/http"
	"time"
)

func UpdateController(metricsService metrics.IMetrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		metricTypeStringAsString := req.PathValue("metricType")
		metricType, err := metrics.ParseMetricType(metricTypeStringAsString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricName := req.PathValue("metricName")
		if len(metricName) == 0 {
			http.Error(w, "empty metric name", http.StatusNotFound)
			return
		}

		metricValue := req.PathValue("metricValue")
		if len(metricValue) == 0 {
			http.Error(w, "empty metric value", http.StatusBadRequest)
			return
		}

		fmt.Println(fmt.Printf("called with metricType='%s', metricName='%s', metricValue='%s' at %s", metricType, metricName, metricValue, time.Now().UTC()))

		err = metricsService.SetMetric(metricType, metricName, metricValue)

		if errors.Is(err, metrics.ErrValueParse) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := metricsService.SetMetric(metricType, metricName, metricValue); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
