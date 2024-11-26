package controllers

import (
	"bytes"
	"github.com/Kopleman/metcol/internal/metrics"
	"net/http"
)

func MainPage(metricsService metrics.IMetrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		allMetrics, err := metricsService.GetAllValuesAsString()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for metricName, metricValue := range allMetrics {
			kvw := bytes.NewBufferString(metricName + ":" + metricValue + "\n")
			if _, err := kvw.WriteTo(w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}
