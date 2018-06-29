package store

import (
	dbm "github.com/tendermint/tmlibs/db"
)

var _ KVStore = transientStore{}

// transientStore is a wrapper for a MemDB with Commiter implementation
type transientStore struct {
	dbStoreAdapter
}

// Constructs new MemDB adapter
func newTransientStore() transientStore {
	return transientStore{dbStoreAdapter{dbm.NewMemDB()}}
}

// Testing purpose
func (ts transientStore) isEmpty() bool {
	return ts.dbStoreAdapter.DB == nil
}
