package main

import (
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/routers"
	"github.com/Kopleman/metcol/internal/store"
	"net/http"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)
	routes := routers.BuildServerRoutes(metricsService)

	return http.ListenAndServe(`:8080`, routes)
}
