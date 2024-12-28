// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package pgxstore

import (
	"context"
)

type Querier interface {
	CreateMetric(ctx context.Context, arg CreateMetricParams) (*Metric, error)
	ExistsMetric(ctx context.Context, arg ExistsMetricParams) (bool, error)
	GetAllMetrics(ctx context.Context) ([]*Metric, error)
	GetMetric(ctx context.Context, arg GetMetricParams) (*Metric, error)
	UpdateMetric(ctx context.Context, arg UpdateMetricParams) error
}

var _ Querier = (*Queries)(nil)
