package postgres

import (
	"context"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgreSQL struct {
	*pgxpool.Pool
}

func NewPostgreSQL(ctx context.Context, logger log.Logger, dsn string) (*PostgreSQL, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	config.MinConns = 3

	config.MaxConns = 6

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	psql := &PostgreSQL{pool}

	if err := psql.PingDB(); err != nil {
		return nil, err
	}

	logger.Info("connected to postgres")

	return psql, nil
}

func (p *PostgreSQL) PingDB() error {
	return p.Ping(context.Background())
}

func (p *PostgreSQL) Interface() IPgxPool {
	return p
}
