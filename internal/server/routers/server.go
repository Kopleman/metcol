package routers

import (
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/gofiber/fiber/v2"
	loggerMW "github.com/gofiber/fiber/v2/middleware/logger"
)

type Metrics interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	SetMetricByDto(metricDto *dto.MetricDTO) error
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(metricType common.MetricType, name string) (*dto.MetricDTO, error)
	GetAllValuesAsString() (map[string]string, error)
}

func BuildAppRoutes(logger log.Logger, app *fiber.App, metricsService Metrics) {
	mainPageCtrl := controllers.NewMainPageController(logger, metricsService)
	updateCtrl := controllers.NewUpdateMetricsController(logger, metricsService)
	getValCtrl := controllers.NewGetValueController(logger, metricsService)

	app.Use(
		loggerMW.New(loggerMW.Config{
			TimeFormat: "2006-01-02T15:04:05.000Z0700",
			TimeZone:   "Local",
			Format:     "${time} | ${status} | ${latency}  | ${method} | ${path} | ${bytesSent} | ${error}\n",
		}),
	)

	apiRouter := app.Group("/")
	app.Get("/", mainPageCtrl.MainPage())

	updateGrp := apiRouter.Group("/update")
	updateGrp.Post("/", updateCtrl.UpdateOrSetViaDTO())
	updateGrp.Post("/:metricType/:metricName/:metricValue", updateCtrl.UpdateOrSet())

	valueGrp := apiRouter.Group("/value")
	valueGrp.Post("/", getValCtrl.GetValueAsDTO())
	valueGrp.Get("/:metricType/:metricName", getValCtrl.GetValue())

}
