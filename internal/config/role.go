package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"revit/internal/db"
	"revit/internal/permissions"
	"strings"

	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name  string             `yaml:"name"`
	Perms []PermissionConfig `yaml:"permissions"`
}

func LoadRolesConfigFromFile(file *os.File) (RoleConfig, error) {
	dec := yaml.NewDecoder(file)
	role := RoleConfig{}
	err := dec.Decode(&role)
	if err != nil {
		return RoleConfig{}, err
	}

	for i := range role.Perms {
		if err := role.Perms[i].ParsePermissionConfig(); err != nil {
			return RoleConfig{}, err
		}
	}

	return role, nil
}

// Config for a permission.
//
// # Permission String Format
// context :: action-allowed
type PermissionConfig struct {
	PermissionStr string                       `yaml:"permission"` // Permissions as the original permission string
	Action        permissions.ActionPermission `yaml:"-"`
	Context       string                       `yaml:"-"`
}

func (p *PermissionConfig) ParsePermissionConfig() error {
	parts := strings.Split(p.PermissionStr, "::")
	if len(parts) != 2 {
		return ConfigError{errType: ConfigErrorInvaidFormatting, msg: fmt.Sprintf(`invalid permission string format,
			found string: %s, with: %d part when split by '::'`, p.PermissionStr, len(parts))}
	}

	p.Context = strings.TrimSpace(parts[0])
	action, err := permissions.ParseActionPermission(strings.TrimSpace(parts[1]))
	if err != nil {
		return err
	}
	p.Action = action
	return nil
}

type RoleConfigStore interface {
	DeleteRoleByName(context.Context, string) error
	GetUnusedRoles(context.Context) ([]db.Role, error)
	ExecTx(context.Context, func(db.Querier) error) error
}

type RoleConfigMgr struct {
	Roles []RoleConfig
	store RoleConfigStore
}

func MakeRoleConfigMgrAndLoad(store RoleConfigStore, configPath string) (*RoleConfigMgr, []forDirResult, error) {
	roleMgr := &RoleConfigMgr{store: store}
	results, err := forDir(configPath, func(file *os.File) error {
		err := roleMgr.LoadRoles(file)
		if err != nil {
			return err
		}
		return nil
	})

	return roleMgr, results, err
}

func (rcm *RoleConfigMgr) LoadRoles(f *os.File) error {
	role, err := LoadRolesConfigFromFile(f)
	if err != nil {
		return err
	}

	rcm.Roles = append(rcm.Roles, role)
	return nil
}

// Sync roles from config. Deletes roles not listed in the config and not asigned to any users.
func (c *RoleConfigMgr) SyncRoles(ctx context.Context, contextsMgr *ContextsConfigMgr) error {
	confNames := make(map[string]struct{})
	for _, role := range c.Roles {
		confNames[role.Name] = struct{}{}
		if err := c.syncRole(ctx, contextsMgr, role); err != nil {
			return err
		}
	}

	rolesArr, err := c.store.GetUnusedRoles(ctx)
	if err != nil {
		return err
	}

	for _, role := range rolesArr {
		if _, ok := confNames[role.Name]; !ok {
			err := c.store.DeleteRoleByName(ctx, role.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Sync a single role.
func (c *RoleConfigMgr) syncRole(ctx context.Context, contextsMgr *ContextsConfigMgr, role RoleConfig) error {
	permissionSets := make(map[string]permissions.PermissionSet, len(role.Perms))
	for _, perm := range role.Perms {
		set, ok := permissionSets[perm.Context]
		if !ok {
			permissionSets[perm.Context] = permissions.PermissionSet{Context: perm.Context, Permissions: map[permissions.ActionPermission]struct{}{perm.Action: struct{}{}}}
		} else {
			set.Permissions[perm.Action] = struct{}{}
			permissionSets[perm.Context] = set
		}
	}
	queryFn := func(q db.Querier) error {
		if err := q.InsertRole(ctx, role.Name); err != nil {
			return err
		}

		ids := []int32{}

		for _, set := range permissionSets {
			if err := permissions.QueriesInsertPermissionSet(ctx, q, set); err != nil {
				return err
			}

			var contextId int64
			if _, found := contextsMgr.Contexts[set.Context]; !found {
				// TODO: print error message about non existent app context
			} else {
				dbContext, err := q.GetContextByName(ctx, set.Context)
				if err != nil {
					return err
				}
				contextId = dbContext.ID
			}
			for action := range set.Permissions {

				id, err := q.GetPermissionId(ctx, db.GetPermissionIdParams{
					Context: contextId,
					Action:  action.String(),
				})

				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						continue
					}
					return err
				}

				ids = append(ids, id)
			}
		}

		dbrole, err := q.GetRoleByName(ctx, role.Name)
		if err != nil {
			return err
		}

		for _, id := range ids {
			if err := q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{
				RoleID:       dbrole.ID,
				PermissionID: id,
			}); err != nil {
				return err
			}
		}

		return nil
	}
	return c.store.ExecTx(ctx, queryFn)
}
