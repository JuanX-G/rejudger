package permissions

import (
	"context"
	db "revit/internal/db"
)

// Store methods for all DB accesses basePermissionManager needs.
type basePermissionMgrStore interface {
	GetUserPermissions(context.Context, int64) ([]db.Permission, error)
	InsertPermission(context.Context, db.InsertPermissionParams) error
	DeletePermission(context.Context, db.DeletePermissionParams) error
	DeletePermissionContext(context.Context, string) error
	InsertRole(context.Context, string) error
	GetRoleByName(context.Context, string) (db.Role, error)
	AddPermissionToRole(context.Context, db.AddPermissionToRoleParams) error
	GetPermissionId(context.Context, db.GetPermissionIdParams) (int32, error)
	ExecTx(context.Context, func(db.Querier) error) error
}
