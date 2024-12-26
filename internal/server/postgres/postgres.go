package postgres

import (
	"context"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	*pgxpool.Pool
}

func NewPostgresSQL(ctx context.Context, logger log.Logger, dsn string) (*PostgreSQL, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("NewPostgresSQL: pgxpool.ParseConfig: %w", err)
	}

	config.MinConns = 3

	config.MaxConns = 6

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("NewPostgresSQL: pool conection error: %w", err)
	}

	psql := &PostgreSQL{pool}

	if pingErr := psql.pingDB(); pingErr != nil {
		return nil, fmt.Errorf("NewPostgresSQL: pingDB: %w", err)
	}

	logger.Info("connected to postgres")

	return psql, nil
}

func (p *PostgreSQL) pingDB() error {
	err := p.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("PingDB: %w", err)
	}

	return nil
}

func (p *PostgreSQL) Interface() IPgxPool {
	return p
}
