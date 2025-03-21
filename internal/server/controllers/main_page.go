package controllers

import (
	"context"
	"net/http"
	"sort"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
)

type MetricsForMainPage interface {
	GetAllValuesAsString(ctx context.Context) (map[string]string, error)
}

// MainPageController instance of controller.
type MainPageController struct {
	logger         log.Logger         // logger
	metricsService MetricsForMainPage // metrics service
}

// NewMainPageController creates instance of controller.
func NewMainPageController(logger log.Logger, metricsService MetricsForMainPage) *MainPageController {
	return &MainPageController{logger, metricsService}
}

// MainPage fetch metric value list
//
//	@MainPage		fetch metric value list
//	@Description	fetch metric value list
//	@Tags			metrics
//	@Produce		plain
//	@Success		200		{string}			"OK"
//	@Failure		500		"Internal Server Error"
//	@Router			/ [get]
func (ctrl *MainPageController) MainPage() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		allMetrics, err := ctrl.metricsService.GetAllValuesAsString(ctx)
		if err != nil {
			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
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
				http.Error(w, common.Err500Message, http.StatusInternalServerError)
				return
			}
			resp += metricName + ":" + metricValue + "\n"
		}

		w.Header().Set(common.ContentType, "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(resp)); err != nil {
			ctrl.logger.Error(err)
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}
	}
}
