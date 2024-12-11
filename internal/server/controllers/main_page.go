package controllers

import (
	"sort"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/gofiber/fiber/v3"
)

type MetricsForMainPage interface {
	GetAllValuesAsString() (map[string]string, error)
}

type MainPageController struct {
	logger         log.Logger
	metricsService MetricsForMainPage
}

func NewMainPageController(logger log.Logger, metricsService MetricsForMainPage) *MainPageController {
	return &MainPageController{logger, metricsService}
}

func (ctrl *MainPageController) MainPage() fiber.Handler {
	return func(c fiber.Ctx) error {
		allMetrics, err := ctrl.metricsService.GetAllValuesAsString()
		if err != nil {
			ctrl.logger.Error(err)
			return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
		}

		var metricNameList []string
		for metricName := range allMetrics {
			metricNameList = append(metricNameList, metricName)
		}
		sort.Strings(metricNameList)

		resp := ""
		for _, metricName := range metricNameList {
			metricValue, ok := allMetrics[metricName]
			if !ok {
				ctrl.logger.Errorf("unable to find metric by name '%s", metricName)
				return fiber.NewError(fiber.StatusInternalServerError, common.Err500Message)
			}
			resp += metricName + ":" + metricValue + "\n"
		}

		return c.SendString(resp)
	}
}
