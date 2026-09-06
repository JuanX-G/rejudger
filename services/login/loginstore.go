package login

import (
	"context"
	"revit/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type loginStore interface {
	GetBasicInfoByUserName(context.Context, string) (db.GetBasicInfoByUserNameRow, error)
	GetBasicInfoByEmail(context.Context, pgtype.Text) (db.GetBasicInfoByEmailRow, error)
}
