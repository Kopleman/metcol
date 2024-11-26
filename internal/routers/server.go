package routers

import (
	"github.com/Kopleman/metcol/internal/controllers"
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/middlewares"
	"github.com/go-chi/chi/v5"
)

func BuildServerRoutes(metricsService metrics.IMetrics) *chi.Mux {
	mainPageCtrl := controllers.MainPage(metricsService)
	updateCtrl := controllers.UpdateController(metricsService)
	getValCtrl := controllers.GetValue(metricsService)

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
