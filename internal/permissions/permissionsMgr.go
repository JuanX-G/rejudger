package permissions

import (
	"context"
	db "revit/internal/db"
	"revit/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionsManager struct {
	db    *pgxpool.Pool
	store *store.Store
	ctx   context.Context
}

func NewPermissionManager(ctx context.Context, pool *pgxpool.Pool) (*PermissionsManager, error) {
	return &PermissionsManager{db: pool, store: store.NewStore(pool), ctx: ctx}, nil
}

func (p *PermissionsManager) GetUserPermissions(userId int) ([]PermissionSet, error) {
	perms, err := p.store.Queries.GetUserPermissions(p.ctx, int64(userId))
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

func (p *PermissionsManager) InsertPermission(ctx context.Context, context string, action ActionPermission) error {
	return p.store.Queries.InsertPermission(ctx, db.InsertPermissionParams{Context: context, Action: action.String()})
}

func (p *PermissionsManager) InsertPermissionSet(ctx context.Context, set PermissionSet) error {
	queryFn := func(q *db.Queries) error {
		for k := range set.Permissions {
			err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return p.store.ExecTx(ctx, queryFn)
}

func QueriesInsertPermissionSet(ctx context.Context, q *db.Queries, set PermissionSet) error {
	for k := range set.Permissions {
		err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PermissionsManager) DeletePermission(ctx context.Context, context string, action ActionPermission) error {
	return p.store.Queries.DeletePermission(ctx, db.DeletePermissionParams{Context: context, Action: action.String()})
}

func (p *PermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	return p.store.Queries.DeletePermissionContext(ctx, context)
}

func (p *PermissionsManager) InsertRole(ctx context.Context, name string) error {
	return p.store.Queries.InsertRole(ctx, name)
}

func (p *PermissionsManager) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	role, err := p.store.Queries.GetRoleByName(ctx, name)
	if err != nil {
		return db.Role{}, err
	}
	return role, nil
}

func (p *PermissionsManager) AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) error {
	return p.store.Queries.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{PermissionID: permissionID, RoleID: roleID})
}

func (p *PermissionsManager) GetPermissionId(ctx context.Context, context string, action ActionPermission) (int, error) {
	id, err := p.store.Queries.GetPermissionId(ctx, db.GetPermissionIdParams{Context: context, Action: action.String()})
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (p *PermissionsManager) GetExecTx() func(context.Context, func(*db.Queries) error) error {
	return p.store.ExecTx
}

func (p *PermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms PermissionSet) error {
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
	return p.store.ExecTx(ctx, queryFn)
}
