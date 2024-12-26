package routers

import (
	"context"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/Kopleman/metcol/internal/server/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Metrics interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	SetMetricByDto(metricDto *dto.MetricDTO) error
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(metricType common.MetricType, name string) (*dto.MetricDTO, error)
	GetAllValuesAsString() (map[string]string, error)
}

type PgxPool interface {
	Ping(context.Context) error
}

func BuildServerRoutes(logger log.Logger, metricsService Metrics, db PgxPool) *chi.Mux {
	mainPageCtrl := controllers.NewMainPageController(logger, metricsService)
	updateCtrl := controllers.NewUpdateMetricsController(logger, metricsService)
	getValCtrl := controllers.NewGetValueController(logger, metricsService)
	pingCtrl := controllers.NewPingController(db)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	// r.Use(middleware.Compress(5, "text/html", "application/json"))
	r.Use(middlewares.CompressMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/", mainPageCtrl.MainPage())
		r.Get("/ping", pingCtrl.Ping())
	})

	r.Route("/update", func(r chi.Router) {
		r.Use(middlewares.PostFilterMiddleware)
		r.Post("/", updateCtrl.UpdateOrSetViaDTO())
		r.Post("/{metricType}/{metricName}/{metricValue}", updateCtrl.UpdateOrSet())
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", getValCtrl.GetValue())
		r.Post("/", getValCtrl.GetValueAsDTO())
	})

	return r
}
