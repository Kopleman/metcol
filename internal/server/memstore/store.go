package memstore

import (
	"context"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/server/store_errors"
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
		return store_errors.ErrAlreadyExists
	}

	s.db[key] = value

	return nil
}

func (s *Store) Read(_ context.Context, mType common.MetricType, name string) (*dto.MetricDTO, error) {
	key := s.buildStoreKey(name, mType)
	value, existed := s.db[key]

	if !existed {
		return nil, store_errors.ErrNotFound
	}

	return value, nil
}

func (s *Store) Update(ctx context.Context, value *dto.MetricDTO) error {
	if _, err := s.Read(ctx, value.MType, value.ID); err != nil {
		return err
	}

	key := s.buildStoreKey(value.ID, value.MType)
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

type Store struct {
	db map[string]*dto.MetricDTO
}

func NewStore(db map[string]*dto.MetricDTO) *Store {
	return &Store{
		db,
	}
}
