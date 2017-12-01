package storage

import (
	"github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

type boltStore struct {
	db *bolt.DB
}

func NewBoltStore(path string) (*boltStore, error) {
	db, err := bolt.Open(path, 0600, bolt.DefaultOptions)
	if err != nil {
		return nil, err
	}
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("images"))
		return nil
	})
	return &boltStore{db}, nil
}

func (b *boltStore) Get(key string) ([]byte, error) {
	var res []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("images"))
		if bucket == nil {
			return errors.New("bucket doesn't exist")
		}

		data := bucket.Get([]byte(key))
		if len(data) == 0 {
			return ErrNotFound
		}

		res = make([]byte, len(data))
		copy(res, data)
		return nil
	})
	return res, err
}

func (b *boltStore) Set(key string, data []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("images"))
		if bucket == nil {
			return errors.New("bucket doesn't exist")
		}
		return bucket.Put([]byte(key), data)
	})
}

func (b *boltStore) ForEach(f func(key string, data []byte)) error {
	return b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("images"))
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			f(string(k), v)
		}
		return nil
	})
}

func (b *boltStore) Close() error {
	return b.db.Close()
}
