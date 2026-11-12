package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"revit/internal/db"
	"revit/internal/jsonHelpers"
	"revit/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfigMgr struct {
	Base        *BaseConfig
	Pipelines   []Pipeline
	RoleMgr     *RoleConfigMgr
	ContextsMgr *ContextsConfigMgr
	store       ConfigStore
}

func NewConfigMgr(pool *pgxpool.Pool, fileName string) (*ConfigMgr, []forDirResult, error) {
	base, err := LoadBaseConfig(fileName)
	if err != nil {
		return nil, []forDirResult{}, err
	}
	if base.RemoteUserManagment == "" || base.RemoteUserManagment == "T" {
		base.RemoteUserManagmentOn = true
	}

	var totalResults []forDirResult

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
	cfgStore := store.NewStore(pool)

	roleMgr, results, err := MakeRoleConfigMgrAndLoad(cfgStore, base.RoleConfigPath)
	totalResults = append(totalResults, results...)
	if err != nil {
		return nil, pipelinesResults, err
	}

	contextsMgr, results, err := MakeContextsMgrAndLoad(cfgStore, base.ContextConfigPath)
	totalResults = append(totalResults, results...)
	if err != nil {
		return nil, pipelinesResults, err
	}

	return &ConfigMgr{Base: base, Pipelines: pipelines, RoleMgr: roleMgr, ContextsMgr: contextsMgr, store: cfgStore}, totalResults, nil
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
	err := c.ContextsMgr.SyncContexts(ctx)
	if err != nil {
		return err
	}
	err = c.RoleMgr.SyncRoles(ctx, c.ContextsMgr)
	if err != nil {
		return err
	}

	err = c.SyncPipelines(ctx)
	if err != nil {
		return err
	}

	return nil
}
