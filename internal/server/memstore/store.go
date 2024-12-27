package memstore

import (
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

func (s *Store) Create(value *dto.MetricDTO) error {
	key := s.buildStoreKey(value.ID, value.MType)
	if s.existed(key) {
		return store_errors.ErrAlreadyExists
	}

	s.db[key] = value

	return nil
}

func (s *Store) Read(key string) (*dto.MetricDTO, error) {
	value, existed := s.db[key]

	if !existed {
		return nil, store_errors.ErrNotFound
	}

	return value, nil
}

func (s *Store) Update(value *dto.MetricDTO) error {
	key := s.buildStoreKey(value.ID, value.MType)
	if _, err := s.Read(key); err != nil {
		return err
	}

	s.db[key] = value

	return nil
}

func (s *Store) Delete(key string) error {
	if _, err := s.Read(key); err != nil {
		return err
	}

	delete(s.db, key)

	return nil
}

func (s *Store) GetAll() map[string]*dto.MetricDTO {
	return s.db
}

type Store struct {
	db map[string]*dto.MetricDTO
}

func NewStore(db map[string]*dto.MetricDTO) *Store {
	return &Store{
		db,
	}
}
