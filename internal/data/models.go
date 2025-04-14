package data

import (
	"context"
	"database/sql"
	"errors"

	"github.com/wisp167/pvz/internal/db"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	PVZv1 PVZModelv1
	PVZ   PVZModel
}

func NewModels(db_ *sql.DB) (Models, error) {
	//queries := db.New(db_) // Initialize queries without prepared statements
	// Or with prepared statements:
	queries, err := db.Prepare(context.Background(), db_)
	if err != nil {
		return Models{}, err
	}

	return Models{
		PVZ: PVZModel{DB: db_, Queries: queries},
	}, nil
}

// Transaction executes a function within a database transaction
func (m *Models) Transaction(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := m.PVZ.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create new queries instance bound to transaction
	txQueries := m.PVZ.Queries.WithTx(tx)

	// Execute the callback
	if err := fn(txQueries); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ReadOnlyTransaction executes a function within a read-only transaction
func (m *Models) ReadOnlyTransaction(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := m.PVZ.DB.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}

	txQueries := m.PVZ.Queries.WithTx(tx)
	if err := fn(txQueries); err != nil {
		tx.Rollback()
		return err
	}

	// Always rollback read-only transactions
	return tx.Rollback()
}
