package controllers

import (
	"errors"
	"strings"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/metrics"
	"github.com/gofiber/fiber/v2"
)

type MetricsForUpdate interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	SetMetricByDto(metricDto *dto.MetricDto) error
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

func (ctrl *UpdateMetricsController) UpdateOrSet() fiber.Handler {
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

		metricValue := strings.ToLower(c.Params("metricValue"))
		if len(metricValue) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "empty metric value")
		}

		ctrl.logger.Infow(
			"metric update called",
			"metricType", metricType,
			"metricName", metricName,
			"metricValue", metricValue,
		)

		err = ctrl.metricsService.SetMetric(metricType, metricName, metricValue)

		if errors.Is(err, metrics.ErrValueParse) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if err != nil {
			ctrl.logger.Error(err)
			return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextPlain)
		return c.SendStatus(fiber.StatusOK)
	}
}

func (ctrl *UpdateMetricsController) UpdateOrSetViaDTO() fiber.Handler {
	return func(c *fiber.Ctx) error {
		metricDto := new(dto.MetricDto)
		if err := c.BodyParser(metricDto); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		ctrl.logger.Infow(
			"metric update called via JSON endpoint",
			"metricType", metricDto.MType,
			"metricName", metricDto.ID,
			"metricValue", metricDto.Value,
			"metricDelta", metricDto.Delta,
		)

		err := ctrl.metricsService.SetMetricByDto(metricDto)

		if errors.Is(err, metrics.ErrValueParse) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if err != nil {
			ctrl.logger.Error(err)
			return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
		}

		return c.JSON(metricDto)
	}
}
