package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"revit/internal/db"
	"revit/internal/jsonHelpers"
	"revit/internal/permissions"
	"revit/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfigMgr struct {
	Base      *BaseConfig
	Pipelines []Pipeline
	Roles     []RoleConfig
	store     ConfigStore
}

func NewConfigMgr(pool *pgxpool.Pool, fileName string) (*ConfigMgr, []forDirResult, error) {
	base, err := LoadBaseConfig(fileName)
	if err != nil {
		return nil, []forDirResult{}, err
	}
	if base.RemoteUserManagment == "" || base.RemoteUserManagment == "T" {
		base.RemoteUserManagmentOn = true
	}

	pipelines := []Pipeline{}
	pipelinesResults, err := forDir(base.PipelineConfigPath, func(file *os.File) error {
		pipeline, err := LoadPipelineConfig(file)
		if err != nil {
			return ConfigError{errType: ConfigErrorDecodeError, msg: fmt.Sprintf("config error occured; yaml.decoder.decode() reported: %s", err.Error())}
		}
		pipelines = append(pipelines, pipeline)
		return nil
	})

	if err != nil {
		return nil, pipelinesResults, err
	}

	roles := []RoleConfig{}
	roleResults, err := forDir(base.RoleConfigPath, func(file *os.File) error {
		role, err := LoadRolesConfigFromFile(file)
		if err != nil {
			return err
		}

		roles = append(roles, role)
		return nil
	})
	totalResults := append(pipelinesResults, roleResults...)
	if err != nil {
		return nil, pipelinesResults, err
	}

	return &ConfigMgr{Base: base, Pipelines: pipelines, Roles: roles, store: store.NewStore(pool)}, totalResults, nil
}

// Sync roles from config. Deletes roles not listed in the config and not asigned to any users.
func (c *ConfigMgr) SyncRoles(ctx context.Context) error {
	confNames := make(map[string]struct{})
	for _, role := range c.Roles {
		confNames[role.Name] = struct{}{}
		if err := c.syncRole(ctx, role); err != nil {
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
func (c *ConfigMgr) syncRole(ctx context.Context, role RoleConfig) error {
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

			for action := range set.Permissions {
				id, err := q.GetPermissionId(ctx, db.GetPermissionIdParams{
					Context: set.Context,
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

// Sync pipeline stage definitions.
func (c *ConfigMgr) syncStageQueriesGetId(ctx context.Context, q db.Querier, stage PipelineStage) (int64, error) {
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
		Final: stage.Final,
	})
	if err != nil {
		return 0, err
	}

	return retStage.ID, nil
}

// Sync pipelines to the DB. Inserts pipelines if a new one has been defined or old one modified.
// Deletes pipelines which are not referenced by any submissions and are not the current state of
// any of the configured pipelines.
func (c *ConfigMgr) SyncPipelines(ctx context.Context) error {
	queryFn := func(q db.Querier) error {
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
			configuredVersions := make([][]byte, 0, len(c.Pipelines))

			for _, p := range c.Pipelines {
				version, err := p.JSON()
				if err != nil {
					return err
				}
				configuredVersions = append(configuredVersions, version)
			}

			unusedPipeline, err := q.GetUnusedPipelines(ctx)
			if err != nil {
				return err
			}

			// Purge unused pipelines. We must use nested for loops as pipelines contain
			// slices so can't be map keys.
			// We do not delete unused an unused pipeline right away as it could be a
			// freshly configured one.
			for _, unused := range unusedPipeline {
				configured := false

				for _, version := range configuredVersions {
					equal, err := jsonHelpers.JSONCompare(version, unused.Version)
					if err != nil {
						return err
					}

					if equal {
						configured = true
						break
					}
				}

				if !configured {
					if err := q.DeletePipelineById(ctx, unused.ID); err != nil {
						return err
					}
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
