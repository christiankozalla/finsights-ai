package eodhd

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Cache struct {
	db *badger.DB
}

func NewCache(path string) (*Cache, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil) // disable noisy logs
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Get(key string, out any) (bool, error) {
	var raw []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			raw = append([]byte(nil), val...)
			return nil
		})
	})
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), bytes).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

func (c *Cache) Close() error {
	return c.db.Close()
}
