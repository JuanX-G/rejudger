package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"revit/internal/db"
	"revit/internal/permissions"
	"revit/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

type ConfigMgr struct {
	Base      *BaseConfig
	Pipelines []Pipeline
	Roles     []RoleConfig
	store     store.Store
}

func NewConfigMgr(pool *pgxpool.Pool, fileName string) (*ConfigMgr, error) {
	base, err := LoadBaseConfig(fileName)
	if err != nil {
		return nil, err
	}
	if base.RemoteUserManagment == "" || base.RemoteUserManagment == "T" {
		base.RemoteUserManagmentOn = true
	}
	pipelines := []Pipeline{}
	forDir(base.PipelieConfigPath, func(file *os.File) error {
		dec := yaml.NewDecoder(file)
		pipeline := Pipeline{}
		err = dec.Decode(&pipeline)
		if err != nil {
			return ConfigError{errType: ConfigErrorDecodeError, msg: fmt.Sprintf("config error occured; yaml.decoder.decode() reported: %s", err.Error())}
		}
		pipelines = append(pipelines, pipeline)
		return nil
	})
	roles := []RoleConfig{}
	forDir(base.RoleConfigPath, func(file *os.File) error {
		role, err := LoadRolesConfigFromFile(file)
		if err != nil {
			return err
		}
		roles = append(roles, role)
		return nil
	})
	return &ConfigMgr{Base: base, Pipelines: pipelines, Roles: roles, store: store.NewStore(pool)}, nil
}

// TODO: add deleting roles connected to a user.
// sync roles from config. Deletes roles not listed in the config and not asigned to any users.
func (c *ConfigMgr) SyncRoles(ctx context.Context) error {
	confNames := make(map[string]struct{})
	for _, role := range c.Roles {
		confNames[role.name] = struct{}{}
		if err := c.syncRole(ctx, role); err != nil {
			return err
		}
	}

	rolesArr, err := c.store.GetQueries().GetUnusedRoles(ctx)
	if err != nil {
		return err
	}

	for _, role := range rolesArr {
		if _, ok := confNames[role.Name]; !ok {
			err := c.store.GetQueries().DeleteRoleByName(ctx, role.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// sync a single role.
func (c *ConfigMgr) syncRole(ctx context.Context, role RoleConfig) error {
	permissionSets := make(map[string]permissions.PermissionSet, len(role.perms))
	for _, perm := range role.perms {
		set, ok := permissionSets[perm.context]
		if !ok {
			permissionSets[perm.context] = permissions.PermissionSet{Context: perm.context, Permissions: map[permissions.ActionPermission]struct{}{perm.action: struct{}{}}}
		} else {
			set.Permissions[perm.action] = struct{}{}
			permissionSets[perm.context] = set
		}
	}
	queryFn := func(q *db.Queries) error {
		if err := q.InsertRole(ctx, role.name); err != nil {
			return nil
		}
		ids := []int32{}
		for _, set := range permissionSets {
			err := permissions.QueriesInsertPermissionSet(ctx, q, set)
			if err != nil {
				return err
			}
			for action := range set.Permissions {
				id, err := q.GetPermissionId(ctx, db.GetPermissionIdParams{Context: set.Context, Action: action.String()})
				if err != nil {
					if !errors.Is(err, pgx.ErrNoRows) {
						continue
					}
					return err
				}
				ids = append(ids, id)
			}
		}
		dbrole, err := q.GetRoleByName(ctx, role.name)
		if err != nil {
			return err
		}
		for _, id := range ids {
			err := q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{RoleID: dbrole.ID, PermissionID: id})
			if err != nil {
				continue
			}
		}
		return nil
	}
	return c.store.ExecTx(ctx, queryFn)
}

// Sync pipeline stage definitions.
func (c *ConfigMgr) syncStageQueriesGetId(ctx context.Context, q *db.Queries, stage PipelineStage) (int64, error) {
	version, err := stage.JSON()
	if err != nil {
		return 0, err
	}

	retStage, err := q.GetStageByVersion(ctx, version)
	if err == nil {
		return retStage.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	retStage, err = q.InsertStage(ctx, db.InsertStageParams{
		Name:            stage.Name,
		Version:         version,
		PlusOneRequired: stage.PlusOnesRequired,
		Blind:           stage.Blind,
		SoftVeto:        stage.SoftVeto,
		HasDeadline:     stage.HasDeadline,
		DeadlineStr: pgtype.Text{
			String: stage.DeadlineStr,
			Valid:  stage.DeadlineStr != "",
		},
	})
	if err != nil {
		return 0, err
	}

	return retStage.ID, nil
}

// Sync pipelines to the DB. Inserts pipelines if a new one has been defined or old one modified.
// Deletes pipelines which are not referenced by any submissions and are not the current state of
// any of the configured pipelines.
// TODO: delete unused pipelines
func (c *ConfigMgr) SyncPipelines(ctx context.Context) error {
	queryFn := func(q *db.Queries) error {
		for _, p := range c.Pipelines {
			version, err := p.JSON()
			if err != nil {
				return err
			}

			_, err = q.GetPipelineByVersion(ctx, version)
			if err == nil {
				continue
			}

			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}

			stageIDs := make([]int64, 0, len(p.Stages))

			for _, stage := range p.Stages {
				stageID, err := c.syncStageQueriesGetId(ctx, q, stage)
				if err != nil {
					return err
				}

				stageIDs = append(stageIDs, stageID)
			}

			pipeline, err := q.InsertPipeline(ctx, db.InsertPipelineParams{
				Name:    p.Name,
				Version: version,
				Description: pgtype.Text{
					String: p.Description,
					Valid:  p.Description != "",
				},
			})
			if err != nil {
				return err
			}

			for position, stageID := range stageIDs {
				err := q.InsertPipelineStage(ctx, db.InsertPipelineStageParams{
					PipelineID: pipeline.ID,
					StageID:    stageID,
					Position:   int32(position),
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return c.store.ExecTx(ctx, queryFn)
}

func (c *ConfigMgr) SyncAll(ctx context.Context) error {
	err := c.SyncRoles(ctx)
	if err != nil {
		return err
	}

	err = c.SyncPipelines(ctx)
	if err != nil {
		return err
	}

	return nil
}
