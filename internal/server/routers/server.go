package routers

import (
	"context"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/controllers"
	"github.com/Kopleman/metcol/internal/server/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Metrics interface {
	SetMetric(ctx context.Context, metricType common.MetricType, name string, value string) error
	SetMetricByDto(ctx context.Context, metricDto *dto.MetricDTO) error
	GetValueAsString(ctx context.Context, metricType common.MetricType, name string) (string, error)
	GetMetricAsDTO(ctx context.Context, metricType common.MetricType, name string) (*dto.MetricDTO, error)
	GetAllValuesAsString(ctx context.Context) (map[string]string, error)
	SetMetrics(ctx context.Context, metrics []*dto.MetricDTO) error
}

type PgxPool interface {
	Ping(context.Context) error
}

func BuildServerRoutes(cfg *config.Config, logger log.Logger, metricsService Metrics, db PgxPool) *chi.Mux {
	mainPageCtrl := controllers.NewMainPageController(logger, metricsService)
	updateCtrl := controllers.NewUpdateMetricsController(logger, metricsService)
	getValCtrl := controllers.NewGetValueController(logger, metricsService)
	pingCtrl := controllers.NewPingController(db)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	// r.Use(middleware.Compress(5, "text/html", "application/json"))
	r.Use(middlewares.CompressMiddleware)
	r.Use(middlewares.Hash(logger, cfg.Key))

	r.Route("/", func(r chi.Router) {
		r.Get("/", mainPageCtrl.MainPage())
		r.Get("/ping", pingCtrl.Ping())
	})

	r.Route("/update", func(r chi.Router) {
		r.Use(middlewares.PostFilterMiddleware)
		r.Post("/", updateCtrl.UpdateOrSetViaDTO())
		r.Post("/{metricType}/{metricName}/{metricValue}", updateCtrl.UpdateOrSet())
	})

	r.Route("/updates", func(r chi.Router) {
		r.Use(middlewares.PostFilterMiddleware)
		r.Post("/", updateCtrl.UpdateMetrics())
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", getValCtrl.GetValue())
		r.Post("/", getValCtrl.GetValueAsDTO())
	})

	return r
}
