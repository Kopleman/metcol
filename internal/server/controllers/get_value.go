package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	errors2 "github.com/Kopleman/metcol/internal/server/errors"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/go-chi/chi/v5"
)

type MetricsForGetValue interface {
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(metricType common.MetricType, name string) (*dto.MetricDTO, error)
}

type GetValueController struct {
	logger         log.Logger
	metricsService MetricsForGetValue
}

func NewGetValueController(logger log.Logger, metricsService MetricsForGetValue) *GetValueController {
	return &GetValueController{logger, metricsService}
}

func (ctrl *GetValueController) GetValue() func(http.ResponseWriter, *http.Request) {
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

		ctrl.logger.Infof("getValue called with metricType='%s', metricName='%s' at %s", metricType, metricName)

		value, err := ctrl.metricsService.GetValueAsString(metricType, metricName)

		if err != nil {
			if errors.Is(err, errors2.ErrNotFound) {
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

func (ctrl *GetValueController) GetValueAsDTO() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		reqDto := new(dto.GetValueRequest)
		if err := json.NewDecoder(req.Body).Decode(&reqDto); err != nil {
			http.Error(w, "unable to parse dto", http.StatusBadRequest)
			return
		}

		ctrl.logger.Infow(
			"get value called via JSON endpoint",
			metricTypeField, reqDto.MType,
			metricNameField, reqDto.ID,
		)

		value, err := ctrl.metricsService.GetMetricAsDTO(reqDto.MType, reqDto.ID)
		if err != nil {
			if errors.Is(err, errors2.ErrNotFound) {
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
