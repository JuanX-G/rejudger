package config

import (
	"context"
	"revit/internal/db"
)

type ConfigStore interface {
	DeleteRoleByName(context.Context, string) error
	GetUnusedRoles(context.Context) ([]db.Role, error)
	ExecTx(context.Context, func(db.Querier) error) error
}
