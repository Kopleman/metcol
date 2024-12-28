package pgxstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/postgres"
	"github.com/Kopleman/metcol/internal/server/sterrors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func (p *PGXStore) StartTx(ctx context.Context, opts *pgx.TxOptions) (pgx.Tx, error) {
	txOpts := pgx.TxOptions{}
	if opts != nil {
		txOpts = *opts
	}
	tx, err := p.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	return tx, nil
}

func (p *PGXStore) WithTx(tx pgx.Tx) *Queries {
	return p.Queries.WithTx(tx)
}

func (p *PGXStore) GetAll(ctx context.Context) ([]*dto.MetricDTO, error) {
	items, err := p.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get metrics from db: %w", err)
	}
	exportData := make([]*dto.MetricDTO, 0, len(items))
	for _, item := range items {
		metricDto, err := item.toDTO()
		if err != nil {
			return nil, fmt.Errorf("could not convert metric to dto: %w", err)
		}
		exportData = append(exportData, metricDto)
	}
	return exportData, nil
}

func (p *PGXStore) commonMetricTypeToPGXMType(mType common.MetricType) (MetricType, error) {
	switch mType {
	case common.CounterMetricType:
		return MetricTypeCounter, nil
	case common.GaugeMetricType:
		return MetricTypeGauge, nil
	default:
		return "", fmt.Errorf("unknown metric type: %v", mType)
	}
}

func (p *PGXStore) Read(ctx context.Context, mType common.MetricType, name string) (*dto.MetricDTO, error) {
	PGXType, err := p.commonMetricTypeToPGXMType(mType)
	if err != nil {
		return nil, fmt.Errorf("could not get metric type for '%s': %w", mType, err)
	}

	item, err := p.GetMetric(ctx, GetMetricParams{
		Type: PGXType,
		Name: name,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sterrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not get metric from db: %w", err)
	}
	if item == nil {
		return nil, sterrors.ErrNotFound
	}

	metricDto, dtoErr := item.toDTO()
	if dtoErr != nil {
		return nil, fmt.Errorf("could not convert metric to dto: %w", err)
	}

	return metricDto, nil
}

func (p *PGXStore) Create(ctx context.Context, metricDTO *dto.MetricDTO) error {
	tx, err := p.StartTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	mType, err := p.commonMetricTypeToPGXMType(metricDTO.MType)
	if err != nil {
		return fmt.Errorf("could not get metric type for '%s': %w", metricDTO.MType, err)
	}
	existed, existErr := p.ExistsMetric(ctx, ExistsMetricParams{
		Name: metricDTO.ID,
		Type: mType,
	})
	if existErr != nil {
		return fmt.Errorf("could not check if metric '%s' exists: %w", metricDTO.ID, existErr)
	}
	if existed {
		return sterrors.ErrAlreadyExists
	}

	createParams := CreateMetricParams{
		Name:  metricDTO.ID,
		Type:  mType,
		Delta: metricDTO.Delta,
		Value: metricDTO.Value,
	}
	_, err = p.CreateMetric(ctx, createParams)
	if err != nil {
		return fmt.Errorf("could not create metric: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

func (p *PGXStore) Update(ctx context.Context, metricDTO *dto.MetricDTO) error {
	tx, err := p.StartTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	mType, err := p.commonMetricTypeToPGXMType(metricDTO.MType)
	if err != nil {
		return fmt.Errorf("could not get metric type for '%s': %w", metricDTO.MType, err)
	}

	existed, existErr := p.ExistsMetric(ctx, ExistsMetricParams{
		Name: metricDTO.ID,
		Type: mType,
	})
	if existErr != nil {
		return fmt.Errorf("could not check if metric '%s' exists: %w", metricDTO.ID, existErr)
	}

	if !existed {
		return p.Create(ctx, metricDTO)
	}

	err = p.UpdateMetric(ctx, UpdateMetricParams{
		Name:  metricDTO.ID,
		Type:  mType,
		Delta: metricDTO.Delta,
		Value: metricDTO.Value,
	})
	if err != nil {
		return fmt.Errorf("could not update metric: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

type PgxPool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Ping(context.Context) error
	Close()
}

type PGXStore struct {
	*Queries
	db     PgxPool
	logger log.Logger
}

func NewPGXStore(l log.Logger, db *postgres.PostgreSQL) *PGXStore {
	return &PGXStore{
		Queries: New(db),
		db:      db,
		logger:  l,
	}
}
