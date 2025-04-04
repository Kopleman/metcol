// Package store interface of required storage.
package store

import (
	"context"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
)

type Store interface {
	Create(ctx context.Context, value *dto.MetricDTO) error
	Read(ctx context.Context, mType common.MetricType, name string) (*dto.MetricDTO, error)
	Update(ctx context.Context, value *dto.MetricDTO) error
	GetAll(ctx context.Context) ([]*dto.MetricDTO, error)
	BulkCreateOrUpdate(ctx context.Context, metricsDTO []*dto.MetricDTO) error
}
