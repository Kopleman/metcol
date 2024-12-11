package routers

import (
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/gofiber/fiber/v3"
)

type Metrics interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetAllValuesAsString() (map[string]string, error)
}

func BuildAppRoutes(logger log.Logger, app *fiber.App, metricsService Metrics) {
	mainPageCtrl := controllers.NewMainPageController(logger, metricsService)
	updateCtrl := controllers.NewUpdateMetricsController(logger, metricsService)
	getValCtrl := controllers.NewGetValueController(logger, metricsService)

	apiRouter := app.Group("/")
	app.Get("/", mainPageCtrl.MainPage())

	updateGrp := apiRouter.Group("/update")
	updateGrp.Post("/:metricType/:metricName/:metricValue", updateCtrl.UpdateOrSet())

	valueGrp := apiRouter.Group("/value")
	valueGrp.Get("/:metricType/:metricName", getValCtrl.GetValue())
}
