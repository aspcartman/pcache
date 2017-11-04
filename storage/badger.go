package storage

import "github.com/dgraph-io/badger"

type badgerStore struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*badgerStore, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &badgerStore{db}, nil
}

func (s *badgerStore) Get(key string) ([]byte, error) {
	var res []byte
	err := s.db.View(func(txn *badger.Txn) error {
		it, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		data, err := it.Value()
		if err != nil {
			return err
		}

		res = data
		return nil
	})
	return res, err
}

func (s *badgerStore) Set(key string, data []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (s *badgerStore) Close() error {
	return s.db.Close()
}
