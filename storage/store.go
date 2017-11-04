package storage

import "github.com/dgraph-io/badger"

var ErrNotFound = badger.ErrKeyNotFound

type Store interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Close() error
}

