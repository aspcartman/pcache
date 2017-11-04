package storage

import "sync"

type inmemStore struct {
	data map[string][]byte
	mtx  *sync.RWMutex
}

func NewInmemStore() *inmemStore {
	return &inmemStore{make(map[string][]byte), &sync.RWMutex{}}
}

func (s *inmemStore) Get(key string) ([]byte, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *inmemStore) Set(key string, data []byte) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.data[key] = data
	return nil
}

func (s *inmemStore) Close() error {
	return nil
}
