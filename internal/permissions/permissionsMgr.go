package permissions

import (
	"context"
	db "revit/internal/db"
)

// Permission manager interface.
type PermissionsManager interface {
	GetUserPermissions(ctx context.Context, userId int) ([]PermissionSet, error)
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

// Default permissionManager implementation.
type BasePermissionsManager struct {
	store basePermissionMgrStore
}

func NewPermissionManager(ctx context.Context, queries basePermissionMgrStore) *BasePermissionsManager {
	return &BasePermissionsManager{store: queries}
}

// Get all permissions, a user with the `userId`, has.
func (p *BasePermissionsManager) GetUserPermissions(ctx context.Context, userId int) ([]PermissionSet, error) {
	perms, err := p.store.GetUserPermissions(ctx, int64(userId))
	if err != nil {
		return []PermissionSet{}, err
	}
	sets := make(map[string]*PermissionSet)

	for _, row := range perms {
		action, err := ParseActionPermission(row.Action)
		if err != nil {
			return nil, err
		}

		appContext, err := p.store.GetContextById(ctx, row.Context)
		if err != nil {
			return nil, err
		}

		set, ok := sets[appContext.Name]
		if !ok {
			set = &PermissionSet{
				Context:     appContext.Name,
				Permissions: make(map[ActionPermission]struct{}),
			}
			sets[appContext.Name] = set
		}

		set.Permissions[action] = struct{}{}
	}

	result := make([]PermissionSet, 0, len(sets))
	for _, set := range sets {
		result = append(result, *set)
	}
	return result, nil
}

func (p *BasePermissionsManager) InsertPermission(ctx context.Context, contextName string, action ActionPermission) error {
	dbcontext, err := p.store.GetContextByName(ctx, contextName)
	if err != nil {
		return err
	}
	return p.store.InsertPermission(ctx, db.InsertPermissionParams{Context: dbcontext.ID, Action: action.String()})
}

func (p *BasePermissionsManager) InsertPermissionSet(ctx context.Context, set PermissionSet) error {
	queryFn := func(q db.Querier) error {
		dbcontext, err := p.store.GetContextByName(ctx, set.Context)
		if err != nil {
			return err
		}

		for k := range set.Permissions {
			err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: dbcontext.ID, Action: k.String()})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return p.store.ExecTx(ctx, queryFn)
}

func (p *BasePermissionsManager) DeletePermission(ctx context.Context, context string, action ActionPermission) error {
	dbcontext, err := p.store.GetContextByName(ctx, context)
	if err != nil {
		return err
	}
	return p.store.DeletePermission(ctx, db.DeletePermissionParams{Context: dbcontext.ID, Action: action.String()})
}

func (p *BasePermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	dbcontext, err := p.store.GetContextByName(ctx, context)
	if err != nil {
		return nil
	}
	return p.store.DeletePermissionContext(ctx, dbcontext.ID)
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
	dbcontext, err := p.store.GetContextByName(ctx, context)
	if err != nil {
		return -1, err
	}

	id, err := p.store.GetPermissionId(ctx, db.GetPermissionIdParams{Context: dbcontext.ID, Action: action.String()})
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (p *BasePermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms PermissionSet) error {
	queryFn := func(q db.Querier) error {
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

		dbcontext, err := p.store.GetContextByName(ctx, perms.Context)
		if err != nil {
			return err
		}

		ids := make([]int32, 0, len(perms.Permissions))
		for k := range perms.Permissions {
			id, err := q.GetPermissionId(ctx, db.GetPermissionIdParams{Context: dbcontext.ID, Action: k.String()})
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
	return p.store.ExecTx(ctx, queryFn)
}
