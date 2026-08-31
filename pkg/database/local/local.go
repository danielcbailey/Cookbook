package local

import (
	"context"
	"sync"

	"github.com/danielcbailey/Cookbook/pkg/database"
)

type localDatabase struct {
	mu    sync.Mutex
	dir   string
	state *store
}

func NewLocal(dir string) (database.Database, error) {
	s, err := loadStore(dir)
	if err != nil {
		return nil, err
	}
	return &localDatabase{dir: dir, state: s}, nil
}

func (db *localDatabase) NewTransaction(_ context.Context) (database.Transaction, error) {
	db.mu.Lock()
	return &localTransaction{db: db, data: db.state.deepCopy()}, nil
}

type localTransaction struct {
	db       *localDatabase
	data     *store
	finished bool
}

func (tx *localTransaction) Commit() error {
	if tx.finished {
		return database.ErrTxClosed
	}
	tx.finished = true

	if err := saveStore(tx.db.dir, tx.data); err != nil {
		tx.db.mu.Unlock()
		return err
	}
	tx.db.state = tx.data
	tx.db.mu.Unlock()
	return nil
}

func (tx *localTransaction) Rollback() error {
	if tx.finished {
		return database.ErrTxClosed
	}
	tx.finished = true
	tx.db.mu.Unlock()
	return nil
}
