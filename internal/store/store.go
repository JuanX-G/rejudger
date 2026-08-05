package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"revit/internal/db"
)

type Store struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: db.New(pool),
	}
}

const MAX_ATTEMPTS = 3

func (s *Store) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
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

func (s *Store) execOnce(ctx context.Context, fn func(*db.Queries) error) error {
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
