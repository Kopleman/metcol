package routers

import (
	"github.com/Kopleman/metcol/internal/metrics"
	controllers2 "github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/Kopleman/metcol/internal/server/middlewares"
	"github.com/go-chi/chi/v5"
)

func BuildServerRoutes(metricsService metrics.IMetrics) *chi.Mux {
	mainPageCtrl := controllers2.MainPage(metricsService)
	updateCtrl := controllers2.UpdateController(metricsService)
	getValCtrl := controllers2.GetValue(metricsService)

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
