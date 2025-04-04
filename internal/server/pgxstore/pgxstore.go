// Package pgxstore implementation of Store interface for postgres storage.
package pgxstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/sterrors"
	"github.com/Kopleman/metcol/internal/server/store"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *PGXStore) StartTx(ctx context.Context) (store.Store, error) {
	tx, err := p.startTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	newQ := p.WithTx(tx)

	return &PGXStore{
		logger:   p.logger,
		Queries:  newQ,
		db:       p.db,
		activeTX: tx,
	}, err
}

func (p *PGXStore) RollbackTx(ctx context.Context) error {
	if p.activeTX == nil {
		return nil
	}

	if err := p.activeTX.Rollback(ctx); err != nil {
		if errors.Is(err, pgx.ErrTxClosed) {
			return nil
		}
		return fmt.Errorf("pgxstore: failed to rollback transaction: %w", err)
	}

	return nil
}

func (p *PGXStore) CommitTx(ctx context.Context) error {
	if p.activeTX == nil {
		return nil
	}
	if err := p.activeTX.Commit(ctx); err != nil {
		return fmt.Errorf("pgxstore: failed to commit transaction: %w", err)
	}
	return nil
}

func (p *PGXStore) startTx(ctx context.Context, opts *pgx.TxOptions) (pgx.Tx, error) {
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
		metricDto, err := item.ToDTO()
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

	item, readErr := p.GetMetric(ctx, GetMetricParams{
		Type: PGXType,
		Name: name,
	})
	if readErr != nil {
		if errors.Is(readErr, pgx.ErrNoRows) {
			return nil, sterrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not get metric from db: %w", readErr)
	}
	if item == nil {
		return nil, sterrors.ErrNotFound
	}

	metricDto, dtoErr := item.ToDTO()
	if dtoErr != nil {
		return nil, fmt.Errorf("could not convert metric to dto: %w", err)
	}

	return metricDto, nil
}

func (p *PGXStore) Create(ctx context.Context, metricDTO *dto.MetricDTO) error {
	mType, err := p.commonMetricTypeToPGXMType(metricDTO.MType)
	if err != nil {
		return fmt.Errorf("could not get metric type for '%s': %w", metricDTO.MType, err)
	}

	createParams := CreateMetricParams{
		Name:  metricDTO.ID,
		Type:  mType,
		Delta: metricDTO.Delta,
		Value: metricDTO.Value,
	}
	_, err = p.CreateMetric(ctx, createParams)
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			return sterrors.ErrAlreadyExists
		}
		return fmt.Errorf("could not create metric: %w", err)
	}

	return nil
}

func (p *PGXStore) Update(ctx context.Context, metricDTO *dto.MetricDTO) error {
	mType, err := p.commonMetricTypeToPGXMType(metricDTO.MType)
	if err != nil {
		return fmt.Errorf("could not get metric type for '%s': %w", metricDTO.MType, err)
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

	return nil
}

func (p *PGXStore) CreateOrUpdate(ctx context.Context, metricDTO *dto.MetricDTO) error {
	mType, err := p.commonMetricTypeToPGXMType(metricDTO.MType)
	if err != nil {
		return fmt.Errorf("pgxstore.CreateOrUpdate type conversion: %w", err)
	}

	_, err = p.CreateOrUpdateMetric(ctx, CreateOrUpdateMetricParams{
		Name:  metricDTO.ID,
		Type:  mType,
		Delta: metricDTO.Delta,
		Value: metricDTO.Value,
	})

	if err != nil {
		return fmt.Errorf("pgxstore.CreateOrUpdate op: %w", err)
	}

	return nil
}

func (p *PGXStore) BulkCreateOrUpdate(ctx context.Context, metricsDTO []*dto.MetricDTO) error {
	tx, err := p.startTx(ctx, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("pgxstore.CreateOrUpdate could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:all // its safe

	for _, metric := range metricsDTO {
		if createOrUpdateErr := p.CreateOrUpdate(ctx, metric); createOrUpdateErr != nil {
			return fmt.Errorf("pgxstore.CreateOrUpdate could not create or update metric: %w", createOrUpdateErr)
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("pgxstore.CreateOrUpdate could not commit transaction: %w", commitErr)
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
	db       PgxPool
	logger   log.Logger
	activeTX pgx.Tx
}

func NewPGXStore(l log.Logger, db PgxPool) *PGXStore {
	return &PGXStore{
		Queries: New(db),
		db:      db,
		logger:  l,
	}
}
