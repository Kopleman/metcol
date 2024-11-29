package controllers

import (
	"bytes"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/metrics"
	"net/http"
)

func MainPage(logger log.Logger, metricsService metrics.IMetrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		allMetrics, err := metricsService.GetAllValuesAsString()
		if err != nil {
			logger.Error(err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		for metricName, metricValue := range allMetrics {
			kvw := bytes.NewBufferString(metricName + ":" + metricValue + "\n")
			if _, err := kvw.WriteTo(w); err != nil {
				logger.Error(err)
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
		}
	}
}
