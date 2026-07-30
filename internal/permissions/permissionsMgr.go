package permissions

import (
	"context"
	db "revit/internal/db"
	"revit/internal/dbUtil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionsManager struct {
	db *pgxpool.Pool
	queries *db.Queries
	ctx context.Context
}

func NewPermissionManager(ctx context.Context, pool *pgxpool.Pool) (*PermissionsManager, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &PermissionsManager{db: pool, queries: db.New(conn), ctx: ctx}, nil
}

func (p *PermissionsManager) GetUserPermissions(userId int) ([]PermissionSet, error) {
	perms, err := p.queries.GetUserPermissions(p.ctx, int64(userId))
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
                Context: row.Context,
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
	return p.queries.InsertPermission(ctx, db.InsertPermissionParams{Context: context, Action: action.String()})
}

func (p *PermissionsManager) InsertPermissionSet(ctx context.Context, set PermissionSet) error {
	tx, _ := p.db.BeginTx(ctx, pgx.TxOptions{})
	q := p.queries.WithTx(tx)
	for k, _ := range set.Permissions {
		err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *PermissionsManager) queriesInsertPermissionSet(ctx context.Context, q *db.Queries, set PermissionSet) error {
	for k, _ := range set.Permissions {
		err := q.InsertPermission(ctx, db.InsertPermissionParams{Context: set.Context, Action: k.String()})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PermissionsManager) DeletePermission(ctx context.Context, context string, action ActionPermission) error {
	return p.queries.DeletePermission(ctx, db.DeletePermissionParams{Context: context, Action: action.String()})
}

func (p *PermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	return p.queries.DeletePermissionContext(ctx, context)
}

func (p *PermissionsManager) InsertRole(ctx context.Context, name string) error {
	return p.queries.InsertRole(ctx, name)
}

func (p *PermissionsManager) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	role, err := p.queries.GetRoleByName(ctx, name)
	if err != nil {
		return db.Role{}, err
	}
	return role, nil
}

func (p *PermissionsManager) AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) (error) {
	return p.queries.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{PermissionID: permissionID, RoleID: roleID})
}


func (p *PermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms PermissionSet) (error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	q := p.queries.WithTx(tx)
	err = q.InsertRole(ctx, roleName)
	if err != nil {
		txErr := tx.Rollback(ctx)
		return dbUtil.DBTxErr(err, txErr)
	}

	role, err := q.GetRoleByName(ctx, roleName)
	if err != nil {
		txErr := tx.Rollback(ctx)
		return dbUtil.DBTxErr(err, txErr)
	}

	err = p.queriesInsertPermissionSet(ctx, q, perms)
	if err != nil {
		txErr := tx.Rollback(ctx)
		return dbUtil.DBTxErr(err, txErr)

	}

	ids := make([]int32, 0, len(perms.Permissions))
	for k, _ := range perms.Permissions {
		id, err := p.queries.GetPermissionId(ctx, db.GetPermissionIdParams{Context: perms.Context, Action: k.String()})
		if err != nil {
			txErr := tx.Rollback(ctx)
			return dbUtil.DBTxErr(err, txErr)

		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		err := q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{RoleID: role.ID, PermissionID: id})
		if err != nil {
			txErr := tx.Rollback(ctx)
			return dbUtil.DBTxErr(err, txErr)

		}
	}
	return nil
}
