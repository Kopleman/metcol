package postgres

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type IPgxPool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Ping(context.Context) error
	Close()
}
