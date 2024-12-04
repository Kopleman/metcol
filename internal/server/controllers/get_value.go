package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi/v5"
)

type MetricsForGetValue interface {
	GetValueAsString(metricType common.MetricType, name string) (string, error)
}

func GetValue(logger log.Logger, metricsService MetricsForGetValue) func(http.ResponseWriter, *http.Request) {
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

		logger.Infof("getValue called with metricType='%s', metricName='%s' at %s", metricType, metricName)

		value, err := metricsService.GetValueAsString(metricType, metricName)
		spew.Dump(err)
		if err != nil {
			spew.Dump(errors.Is(err, store.ErrNotFound))
			if errors.Is(err, store.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		if _, err := io.WriteString(w, value); err != nil {
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
		}
	}
}
