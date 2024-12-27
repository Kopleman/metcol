package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/go-chi/chi/v5"
)

type MetricsForUpdate interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	SetMetricByDto(metricDto *dto.MetricDTO) error
}

type UpdateMetricsController struct {
	logger         log.Logger
	metricsService MetricsForUpdate
}

func NewUpdateMetricsController(logger log.Logger, metricsService MetricsForUpdate) UpdateMetricsController {
	return UpdateMetricsController{
		logger:         logger,
		metricsService: metricsService,
	}
}

func (ctrl *UpdateMetricsController) UpdateOrSet() func(http.ResponseWriter, *http.Request) {
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

		metricValue := strings.ToLower(chi.URLParam(req, "metricValue"))
		if len(metricValue) == 0 {
			http.Error(w, "empty metric value", http.StatusBadRequest)
			return
		}

		ctrl.logger.Infof(
			"update called with metricType='%s', metricName='%s', metricValue='%s'",
			metricType,
			metricName,
			metricValue,
		)

		err = ctrl.metricsService.SetMetric(metricType, metricName, metricValue)

		if errors.Is(err, metrics.ErrValueParse) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}
		w.Header().Set(common.ContentType, "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func (ctrl *UpdateMetricsController) UpdateOrSetViaDTO() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		metricDto := new(dto.MetricDTO)
		if err := json.NewDecoder(req.Body).Decode(&metricDto); err != nil {
			http.Error(w, "unable to parse dto", http.StatusBadRequest)
			return
		}

		ctrl.logger.Infow(
			"metric update called via JSON endpoint",
			metricTypeField, metricDto.MType,
			metricNameField, metricDto.ID,
			metricValueField, metricDto.Value,
			"metricDelta", metricDto.Delta,
		)

		err := ctrl.metricsService.SetMetricByDto(metricDto)

		if errors.Is(err, metrics.ErrValueParse) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusBadRequest)
			return
		}

		w.Header().Set(common.ContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(metricDto); err != nil {
			http.Error(w, common.Err500Message, http.StatusBadRequest)
			return
		}
	}
}
