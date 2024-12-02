package controllers

import (
	"bytes"
	"net/http"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/metrics"
)

func MainPage(logger log.Logger, metricsService metrics.IMetrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		allMetrics, err := metricsService.GetAllValuesAsString()
		if err != nil {
			logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		for metricName, metricValue := range allMetrics {
			kvw := bytes.NewBufferString(metricName + ":" + metricValue + "\n")
			if _, err := kvw.WriteTo(w); err != nil {
				logger.Error(err)
				http.Error(w, common.Err500Message, http.StatusInternalServerError)
				return
			}
		}
	}
}
