package controllers

import (
	"bytes"
	"net/http"
	"sort"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
)

type MetricsForMainPage interface {
	GetAllValuesAsString() (map[string]string, error)
}

func MainPage(logger log.Logger, metricsService MetricsForMainPage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		allMetrics, err := metricsService.GetAllValuesAsString()
		if err != nil {
			logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		var metricNameList []string
		for metricName, _ := range allMetrics {
			metricNameList = append(metricNameList, metricName)
		}
		sort.Strings(metricNameList)
		for _, metricName := range metricNameList {
			metricValue, ok := allMetrics[metricName]
			if !ok {
				logger.Errorf("unable to find metrcia by name '%s", metricName)
				http.Error(w, common.Err500Message, http.StatusInternalServerError)
				return
			}

			kvw := bytes.NewBufferString(metricName + ":" + metricValue + "\n")
			if _, err := kvw.WriteTo(w); err != nil {
				logger.Error(err)
				http.Error(w, common.Err500Message, http.StatusInternalServerError)
				return
			}
		}
	}
}
