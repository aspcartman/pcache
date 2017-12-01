package storage

import (
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

type badgerStore struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*badgerStore, error) {
	opts := badger.DefaultOptions
	opts.TableLoadingMode = options.MemoryMap // map the index instead of copying to save memory usage
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

		res = make([]byte, len(data))
		copy(res, data)
		return nil
	})
	return res, err
}

func (s *badgerStore) Set(key string, data []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (s *badgerStore) ForEach(f func(key string, data []byte)) error {
	return s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			f(string(item.Key()), v)
		}
		return nil
	})
}

func (s *badgerStore) Close() error {
	return s.db.Close()
}
