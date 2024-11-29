package main

import (
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/metrics"
	"github.com/Kopleman/metcol/internal/server/config"
	"github.com/Kopleman/metcol/internal/server/routers"
	"github.com/Kopleman/metcol/internal/server/store"
	"net/http"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	logger := log.New(
		log.WithAppVersion("local"),
	)

	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run(_ log.Logger) error {
	srvConfig, err := config.ParseServerConfig()
	if err != nil {
		return err
	}

	storeService := store.NewStore(make(map[string]any))
	metricsService := metrics.NewMetrics(storeService)
	routes := routers.BuildServerRoutes(metricsService)

	return http.ListenAndServe(srvConfig.NetAddr.String(), routes)
}
