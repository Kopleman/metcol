package store

func (s *Store) existed(key string) bool {
	_, existed := s.db[key]
	return existed
}

func (s *Store) Create(key string, value any) error {
	if s.existed(key) {
		return ErrAlreadyExists
	}

	s.db[key] = value

	return nil
}

func (s *Store) Read(key string) (any, error) {
	value, existed := s.db[key]

	if !existed {
		return nil, ErrNotFound
	}

	return value, nil
}

func (s *Store) Update(key string, value any) error {
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

func (s *Store) GetAll() map[string]any {
	return s.db
}

type Store struct {
	db map[string]any
}

func NewStore(db map[string]any) *Store {
	return &Store{
		db,
	}
}
