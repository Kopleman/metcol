package routers

import (
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/Kopleman/metcol/internal/server/middlewares"
	"github.com/go-chi/chi/v5"
)

type Metrics interface {
	SetMetric(metricType common.MetricType, name string, value string) error
	GetValueAsString(metricType common.MetricType, name string) (string, error)
	GetAllValuesAsString() (map[string]string, error)
}

func BuildServerRoutes(logger log.Logger, metricsService Metrics) *chi.Mux {
	mainPageCtrl := controllers.MainPage(logger, metricsService)
	updateCtrl := controllers.UpdateController(logger, metricsService)
	getValCtrl := controllers.GetValue(logger, metricsService)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", mainPageCtrl)
	})

	r.Route("/update", func(r chi.Router) {
		r.Use(middlewares.PostFilterMiddleware)
		r.Post("/{metricType}/{metricName}/{metricValue}", updateCtrl)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", getValCtrl)
	})

	return r
}
