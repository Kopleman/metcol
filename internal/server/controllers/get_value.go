package controllers

import (
	"errors"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/gofiber/fiber/v2"
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

func (ctrl *GetValueController) GetValue() fiber.Handler {
	return func(c *fiber.Ctx) error {
		metricTypeStringAsString := strings.ToLower(c.Params("metricType"))
		metricType, err := metrics.ParseMetricType(metricTypeStringAsString)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		metricName := strings.ToLower(c.Params("metricName"))
		if len(metricName) == 0 {
			return fiber.NewError(fiber.StatusNotFound, "empty metric name")
		}

		ctrl.logger.Infof("getValue called with metricType='%s', metricName='%s' at %s", metricType, metricName)

		value, err := ctrl.metricsService.GetValueAsString(metricType, metricName)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, err.Error())
			}

			ctrl.logger.Error(err)
			return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
		}

		return c.SendString(value)
	}
}

func (ctrl *GetValueController) GetValueAsDTO() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqDto := new(dto.GetValueRequest)
		if err := c.BodyParser(reqDto); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "unable to parse dto")
		}

		ctrl.logger.Infow(
			"get value called via JSON endpoint",
			"metricType", reqDto.MType,
			"metricName", reqDto.ID,
		)

		value, err := ctrl.metricsService.GetMetricAsDTO(reqDto.MType, reqDto.ID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, err.Error())
			}

			ctrl.logger.Error(err)
			return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
		}

		return c.JSON(value)
	}
}
