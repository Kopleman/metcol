package controllers

import (
	"context"
	"net/http"

	"github.com/Kopleman/metcol/internal/common"
)

type PgxPool interface {
	Ping(context.Context) error
}

type PingController struct {
	db PgxPool
}

func NewPingController(db PgxPool) *PingController {
	return &PingController{db: db}
}

func (ctrl *PingController) Ping() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if err := ctrl.db.Ping(ctx); err != nil {
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		w.Header().Set(common.ContentType, "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
