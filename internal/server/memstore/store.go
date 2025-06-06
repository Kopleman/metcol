// Package memstore implementation of Store interface for im-memo storage.
package memstore

import (
	"context"
	"sync"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/server/sterrors"
)

func (s *Store) buildStoreKey(name string, metricType common.MetricType) string {
	return name + "-" + string(metricType)
}

func (s *Store) existed(key string) bool {
	_, existed := s.db[key]
	return existed
}

func (s *Store) Create(_ context.Context, value *dto.MetricDTO) error {
	key := s.buildStoreKey(value.ID, value.MType)
	if s.existed(key) {
		return sterrors.ErrAlreadyExists
	}

	s.db[key] = value

	return nil
}

func (s *Store) Read(_ context.Context, mType common.MetricType, name string) (*dto.MetricDTO, error) {
	key := s.buildStoreKey(name, mType)
	s.mu.Lock()
	defer s.mu.Unlock()
	value, existed := s.db[key]

	if !existed {
		return nil, sterrors.ErrNotFound
	}

	return value, nil
}

func (s *Store) Update(ctx context.Context, value *dto.MetricDTO) error {
	if _, err := s.Read(ctx, value.MType, value.ID); err != nil {
		return err
	}

	key := s.buildStoreKey(value.ID, value.MType)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db[key] = value

	return nil
}

func (s *Store) Delete(ctx context.Context, mType common.MetricType, name string) error {
	key := s.buildStoreKey(name, mType)
	if _, err := s.Read(ctx, mType, name); err != nil {
		return err
	}

	delete(s.db, key)

	return nil
}

func (s *Store) GetAll(_ context.Context) ([]*dto.MetricDTO, error) {
	exportData := make([]*dto.MetricDTO, 0, len(s.db))
	for _, metricValue := range s.db {
		exportData = append(exportData, metricValue)
	}
	return exportData, nil
}

func (s *Store) BulkCreateOrUpdate(_ context.Context, metricsDTO []*dto.MetricDTO) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metricsDTO {
		key := s.buildStoreKey(metric.ID, metric.MType)
		s.db[key] = metric
	}

	return nil
}

type Store struct {
	db map[string]*dto.MetricDTO
	mu *sync.Mutex
}

func NewStore(db map[string]*dto.MetricDTO) *Store {
	return &Store{
		db: db,
		mu: &sync.Mutex{},
	}
}
