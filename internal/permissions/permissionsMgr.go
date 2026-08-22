package permissions

import (
	"context"
	db "revit/internal/db"
	"revit/internal/store"
)

type PermissionsManager interface {
	GetUserPermissions(userId int) ([]PermissionSet, error)
	InsertPermission(ctx context.Context, context string, action ActionPermission) error
	InsertPermissionSet(ctx context.Context, set PermissionSet) error
	DeletePermission(ctx context.Context, context string, action ActionPermission) error
	DeletePermissionContext(ctx context.Context, context string) error
	InsertRole(ctx context.Context, name string) error
	GetRoleByName(ctx context.Context, name string) (db.Role, error)

	AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) error
	GetPermissionId(ctx context.Context, context string, action ActionPermission) (int, error)
	AddRoleWithPermissions(ctx context.Context, roleName string, perms PermissionSet) error
}

type basePermissionMgrStore interface {
	GetUserPermissions(context.Context, int64) ([]db.Permission, error)
	InsertPermission(context.Context, db.InsertPermissionParams) error
	DeletePermission(context.Context, db.DeletePermissionParams) error
	DeletePermissionContext(context.Context, string) error
	InsertRole(context.Context, string) error
	GetRoleByName(context.Context, string) (db.Role, error)
	AddPermissionToRole(context.Context, db.AddPermissionToRoleParams) error
	GetPermissionId(context.Context, db.GetPermissionIdParams) (int32, error)
}

type BasePermissionsManager struct {
	txstore store.Store
	store   basePermissionMgrStore
	ctx     context.Context
}

func NewPermissionManager(ctx context.Context, queries *db.Queries) (*BasePermissionsManager, error) {
	return &BasePermissionsManager{store: queries, ctx: ctx}, nil
}

func (p *BasePermissionsManager) GetUserPermissions(userId int) ([]PermissionSet, error) {
	perms, err := p.store.GetUserPermissions(p.ctx, int64(userId))
	if err != nil {
		return []PermissionSet{}, err
	}
	sets := make(map[string]*PermissionSet)

	for _, row := range perms {
		action, err := ParseActionPermission(row.Action)
		if err != nil {
			return nil, err
		}

		set, ok := sets[row.Context]
		if !ok {
			set = &PermissionSet{
				Context:     row.Context,
				Permissions: make(map[ActionPermission]struct{}),
			}
			sets[row.Context] = set
		}

		set.Permissions[action] = struct{}{}
	}

	result := make([]PermissionSet, 0, len(sets))
	for _, set := range sets {
		result = append(result, *set)
	}
	return result, nil
}

func (p *BasePermissionsManager) InsertPermission(ctx context.Context, context string, action ActionPermission) error {
	return p.store.InsertPermission(ctx, db.InsertPermissionParams{Context: context, Action: action.String()})
}

func (p *BasePermissionsManager) InsertPermissionSet(ctx context.Context, set PermissionSet) error {
	queryFn := func(q *db.Queries) error {
		for k := range set.Permissions {
			err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return p.txstore.ExecTx(ctx, queryFn)
}

func (p *BasePermissionsManager) DeletePermission(ctx context.Context, context string, action ActionPermission) error {
	return p.store.DeletePermission(ctx, db.DeletePermissionParams{Context: context, Action: action.String()})
}

func (p *BasePermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	return p.store.DeletePermissionContext(ctx, context)
}

func (p *BasePermissionsManager) InsertRole(ctx context.Context, name string) error {
	return p.store.InsertRole(ctx, name)
}

func (p *BasePermissionsManager) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	role, err := p.store.GetRoleByName(ctx, name)
	if err != nil {
		return db.Role{}, err
	}
	return role, nil
}

func (p *BasePermissionsManager) AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) error {
	return p.store.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{PermissionID: permissionID, RoleID: roleID})
}

func (p *BasePermissionsManager) GetPermissionId(ctx context.Context, context string, action ActionPermission) (int, error) {
	id, err := p.store.GetPermissionId(ctx, db.GetPermissionIdParams{Context: context, Action: action.String()})
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (p *BasePermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms PermissionSet) error {
	queryFn := func(q *db.Queries) error {
		err := q.InsertRole(ctx, roleName)
		if err != nil {
			return err
		}

		role, err := q.GetRoleByName(ctx, roleName)
		if err != nil {
			return err
		}

		err = QueriesInsertPermissionSet(ctx, q, perms)
		if err != nil {
			return err
		}

		ids := make([]int32, 0, len(perms.Permissions))
		for k := range perms.Permissions {
			id, err := q.GetPermissionId(ctx, db.GetPermissionIdParams{Context: perms.Context, Action: k.String()})
			if err != nil {
				return err

			}
			ids = append(ids, id)
		}
		for _, id := range ids {
			err := q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{RoleID: role.ID, PermissionID: id})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return p.txstore.ExecTx(ctx, queryFn)
}
