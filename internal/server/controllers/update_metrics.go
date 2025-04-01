package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/go-chi/chi/v5"
)

type MetricsForUpdate interface {
	SetMetric(ctx context.Context, metricType common.MetricType, name string, value string) error
	SetMetricByDto(ctx context.Context, metricDto *dto.MetricDTO) error
	SetMetrics(ctx context.Context, metrics []*dto.MetricDTO) error
}

// UpdateMetricsController instance of controller.
type UpdateMetricsController struct {
	logger         log.Logger       // logger
	metricsService MetricsForUpdate // metrics service
}

// NewUpdateMetricsController creates instance of controller.
func NewUpdateMetricsController(logger log.Logger, metricsService MetricsForUpdate) UpdateMetricsController {
	return UpdateMetricsController{
		logger:         logger,
		metricsService: metricsService,
	}
}

// UpdateOrSet sets metric value
//
//	@Summary		sets metric value
//	@Description	sets metric value if it exists or creates new one
//	@Tags			metrics
//	@Produce		plain
//	@Param			metricType	path		string	true	"Metric type"
//	@Param			metricName	path		string	true	"Metric name"
//	@Param			metricValue	path		string	true	"Metric value"
//	@Success		200		{string}			"OK"
//	@Failure		400		"Bad request"
//	@Failure		404		"Not found"
//	@Failure		500		"Internal Server Error"
//	@Router			/update/{metricType}/{metricName}/{metricValue} [post]
func (ctrl *UpdateMetricsController) UpdateOrSet() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
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

		err = ctrl.metricsService.SetMetric(ctx, metricType, metricName, metricValue)

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

// UpdateOrSetViaDTO sets metric value
//
//	@Summary		sets metric value
//	@Description	sets metric value if it exists or creates new one
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			data			body	dto.MetricDTO	true	"Body params"
//	@Success		200				{object}	dto.MetricDTO
//	@Failure		400		"Bad request"
//	@Failure		500		"Internal Server Error"
//	@Router			/update [post]
func (ctrl *UpdateMetricsController) UpdateOrSetViaDTO() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		metricDto := new(dto.MetricDTO)
		if err := json.NewDecoder(req.Body).Decode(&metricDto); err != nil {
			ctrl.logger.Error(err)
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

		err := ctrl.metricsService.SetMetricByDto(ctx, metricDto)

		if errors.Is(err, metrics.ErrValueParse) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		w.Header().Set(common.ContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(metricDto); err != nil {
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}
	}
}

func (ctrl *UpdateMetricsController) parseUpdateBody(req *http.Request) ([]*dto.MetricDTO, error) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}
	metricsBatch := make([]*dto.MetricDTO, 0)
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&metricsBatch)
	if err == nil {
		return metricsBatch, nil
	}

	metricDto := new(dto.MetricDTO)
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&metricDto); err != nil {
		return nil, fmt.Errorf("unable to parse dto: %w", err)
	}
	metricsBatch = append(metricsBatch, metricDto)
	return metricsBatch, nil
}

// UpdateMetrics sets metrics value
//
//	@Summary		sets metrics value
//	@Description	sets metrics value via bulk
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			data			body	[]dto.MetricDTO	true	"Body params"
//	@Success		200				{array}	dto.MetricDTO
//	@Failure		400		"Bad request"
//	@Failure		500		"Internal Server Error"
//	@Router			/update [post]
func (ctrl *UpdateMetricsController) UpdateMetrics() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		metricsBatch, err := ctrl.parseUpdateBody(req)
		if err != nil {
			ctrl.logger.Error(err)
			http.Error(w, "unable to parse dto", http.StatusBadRequest)
			return
		}

		ctrl.logger.Infow(
			"metrics update called",
			"amount", len(metricsBatch),
		)

		setError := ctrl.metricsService.SetMetrics(ctx, metricsBatch)

		if setError != nil {
			msg := common.Err500Message
			if errors.Is(setError, metrics.ErrValueParse) {
				msg = setError.Error()
			}
			ctrl.logger.Error(setError)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		w.Header().Set(common.ContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(metricsBatch); err != nil {
			http.Error(w, common.Err500Message, http.StatusBadRequest)
			return
		}
	}
}
