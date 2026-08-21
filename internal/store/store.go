package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"revit/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	ExecTx(context.Context, func(db *db.Queries) error) error
	GetQueries() *db.Queries
}

type PgxStore struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewStore(pool *pgxpool.Pool) *PgxStore {
	return &PgxStore{
		pool:    pool,
		Queries: db.New(pool),
	}
}

const MAX_ATTEMPTS = 3

func (s *PgxStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	var lastErr error

	for attempt := 1; attempt <= MAX_ATTEMPTS; attempt++ {
		lastErr = s.execOnce(ctx, fn)
		if lastErr == nil {
			return nil
		}

		var storeErr *StoreError
		if errors.As(lastErr, &storeErr) && storeErr.Retryable() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt*25) * time.Millisecond):
				continue
			}
		}

		return lastErr
	}

	return fmt.Errorf("exceeded max retry attempts (%d): %w", MAX_ATTEMPTS, lastErr)
}

func (s *PgxStore) GetQueries() *db.Queries {
	return s.Queries
}

func (s *PgxStore) execOnce(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return wrapStoreError(err)
	}
	defer tx.Rollback(ctx)

	qtx := db.New(tx)

	if err := fn(qtx); err != nil {
		return wrapStoreError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return wrapStoreError(err)
	}

	return nil
}
