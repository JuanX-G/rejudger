package dbUtil

import (
	"fmt"
)


type DBError struct {
	TxError error
	DBError error
}

func (e DBError) Error() string {
	if e.TxError == nil && e.DBError != nil {
		return fmt.Sprintf("db error: %s", e.DBError.Error())
	} else if e.DBError == nil && e.TxError != nil {
		return fmt.Sprintf("tx error: %s", e.TxError.Error())
	} else {
		return fmt.Sprintf("db error: %s, tx error: %s", e.DBError.Error(), e.TxError.Error())
	}
}

func TxErr(err error) error {
	return DBError{TxError: err}
}

func DBErr(err error) error {
	return DBError{DBError: err}
}

func DBTxErr(DBErr error, TxErr error) error {
	return DBError{DBError: DBErr, TxError: TxErr}
}
