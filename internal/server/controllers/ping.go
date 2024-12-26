package controllers

import (
	"net/http"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/server/postgres"
)

type PingController struct {
	db *postgres.PostgreSQL
}

func NewPingController(db *postgres.PostgreSQL) *PingController {
	return &PingController{db: db}
}

func (ctrl *PingController) Ping() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := ctrl.db.PingDB(); err != nil {
			http.Error(w, common.Err500Message, http.StatusInternalServerError)
			return
		}

		w.Header().Set(common.ContentType, "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
