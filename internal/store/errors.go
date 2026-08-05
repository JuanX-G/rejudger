package store

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type StoreError struct {
	Code      string
	Message   string
	retryable bool
	Err       error
}

func (s *StoreError) Error() string {
	return fmt.Sprintf("db error [%s]: %s", s.Code, s.Message)
}

func (s *StoreError) Unwrap() error {
	return s.Err
}

func (s *StoreError) Retryable() bool {
	return s.retryable
}

func wrapStoreError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code := pgErr.Code
		isRetryable := code == pgerrcode.SerializationFailure || code == pgerrcode.DeadlockDetected

		return &StoreError{
			Code:      code,
			Message:   pgErr.Message,
			retryable: isRetryable,
			Err:       err,
		}
	}

	return err
}
