package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/sterrors"
	"github.com/go-chi/chi/v5"
)

type MetricsForGetValue interface {
	GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error)
}

// GetValueController instance of controller.
type GetValueController struct {
	logger         log.Logger         // logger
	metricsService MetricsForGetValue // metrics service
}

// NewGetValueController creates instance of controller.
func NewGetValueController(logger log.Logger, metricsService MetricsForGetValue) *GetValueController {
	return &GetValueController{logger, metricsService}
}

// GetValue fetch metric value
//
//	@Summary		fetch metric value
//	@Description	fetch metric value
//	@Tags			metrics
//	@Accept			plain
//	@Produce		plain
//	@Param			metricType	path		string	true	"Metric type"
//	@Param			metricName	path		string	true	"Metric name"
//	@Success		200		{string}			"OK"
//	@Failure		400		"Bad request"
//	@Failure		404		"Not found"
//	@Failure		500		"Internal Server Error"
//	@Router			/value/{metricType}/{metricName} [get]
func (ctrl *GetValueController) GetValue() func(http.ResponseWriter, *http.Request) {
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

		ctrl.logger.Infof("getValue called with metricType='%s', metricName='%s' at %s", metricType, metricName)

		value, err := ctrl.metricsService.GetValueAsString(ctx, metricType, metricName)

		if err != nil {
			if errors.Is(err, sterrors.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		w.Header().Set(common.ContentType, "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(value)); err != nil {
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
		}
	}
}

// GetValueAsDTO fetch metric value
//
//	@Summary		fetch metric value
//	@Description	fetch metric value
//	@Tags			metrics
//	@Accept			json
//	@Produce		plain
//	@Param			data			body	dto.GetValueRequest	true	"Body params"
//	@Success		200				{object}	dto.MetricDTO
//	@Failure		400		"Bad request"
//	@Failure		404		"Not found"
//	@Failure		500		"Internal Server Error"
//	@Router			/value [post]
func (ctrl *GetValueController) GetValueAsDTO() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		reqDto := new(dto.GetValueRequest)
		if err := json.NewDecoder(req.Body).Decode(&reqDto); err != nil {
			http.Error(w, common.ErrDtoParse, http.StatusBadRequest)
			return
		}

		ctrl.logger.Infow(
			"get value called via JSON endpoint",
			metricTypeField, reqDto.MType,
			metricNameField, reqDto.ID,
		)

		value, err := ctrl.metricsService.GetMetricAsDTO(ctx, reqDto.MType, reqDto.ID)
		if err != nil {
			if errors.Is(err, sterrors.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		w.Header().Set(common.ContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(value); err != nil {
			http.Error(w, common.Err500Message, http.StatusBadRequest)
			return
		}
	}
}
