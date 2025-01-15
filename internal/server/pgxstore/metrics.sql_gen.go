// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: metrics.sql

package pgxstore

import (
	"context"
)

const CreateMetric = `-- name: CreateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
VALUES ($1, $2, $3, $4, now())
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at
`

type CreateMetricParams struct {
	Value *float64   `db:"value" json:"value"`
	Delta *int64     `db:"delta" json:"delta"`
	Name  string     `db:"name" json:"name"`
	Type  MetricType `db:"type" json:"type"`
}

func (q *Queries) CreateMetric(ctx context.Context, arg CreateMetricParams) (*Metric, error) {
	row := q.db.QueryRow(ctx, CreateMetric,
		arg.Name,
		arg.Type,
		arg.Value,
		arg.Delta,
	)
	var i Metric
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Type,
		&i.Value,
		&i.Delta,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}

const CreateOrUpdateMetric = `-- name: CreateOrUpdateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
VALUES ($1, $2, $3, $4, now())
    ON CONFLICT ON CONSTRAINT name_type_uniq DO UPDATE SET value=$3, delta=$4, updated_at=now()
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at
`

type CreateOrUpdateMetricParams struct {
	Value *float64   `db:"value" json:"value"`
	Delta *int64     `db:"delta" json:"delta"`
	Name  string     `db:"name" json:"name"`
	Type  MetricType `db:"type" json:"type"`
}

func (q *Queries) CreateOrUpdateMetric(ctx context.Context, arg CreateOrUpdateMetricParams) (*Metric, error) {
	row := q.db.QueryRow(ctx, CreateOrUpdateMetric,
		arg.Name,
		arg.Type,
		arg.Value,
		arg.Delta,
	)
	var i Metric
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Type,
		&i.Value,
		&i.Delta,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}

const ExistsMetric = `-- name: ExistsMetric :one
SELECT EXISTS (SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics WHERE name=$1 AND type=$2)::boolean
`

type ExistsMetricParams struct {
	Name string     `db:"name" json:"name"`
	Type MetricType `db:"type" json:"type"`
}

func (q *Queries) ExistsMetric(ctx context.Context, arg ExistsMetricParams) (bool, error) {
	row := q.db.QueryRow(ctx, ExistsMetric, arg.Name, arg.Type)
	var column_1 bool
	err := row.Scan(&column_1)
	return column_1, err
}

const GetAllMetrics = `-- name: GetAllMetrics :many
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics ORDER BY name ASC
`

func (q *Queries) GetAllMetrics(ctx context.Context) ([]*Metric, error) {
	rows, err := q.db.Query(ctx, GetAllMetrics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Metric{}
	for rows.Next() {
		var i Metric
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Type,
			&i.Value,
			&i.Delta,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetMetric = `-- name: GetMetric :one
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics WHERE type=$1 AND name=$2 LIMIT 1
`

type GetMetricParams struct {
	Type MetricType `db:"type" json:"type"`
	Name string     `db:"name" json:"name"`
}

func (q *Queries) GetMetric(ctx context.Context, arg GetMetricParams) (*Metric, error) {
	row := q.db.QueryRow(ctx, GetMetric, arg.Type, arg.Name)
	var i Metric
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Type,
		&i.Value,
		&i.Delta,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}

const UpdateMetric = `-- name: UpdateMetric :one
UPDATE metrics
SET value=$1, delta=$2, updated_at=now()
WHERE type=$3 AND name=$4
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at
`

type UpdateMetricParams struct {
	Value *float64   `db:"value" json:"value"`
	Delta *int64     `db:"delta" json:"delta"`
	Type  MetricType `db:"type" json:"type"`
	Name  string     `db:"name" json:"name"`
}

func (q *Queries) UpdateMetric(ctx context.Context, arg UpdateMetricParams) (*Metric, error) {
	row := q.db.QueryRow(ctx, UpdateMetric,
		arg.Value,
		arg.Delta,
		arg.Type,
		arg.Name,
	)
	var i Metric
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Type,
		&i.Value,
		&i.Delta,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return &i, err
}
