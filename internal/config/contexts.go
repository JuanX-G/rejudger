package config

import (
	"context"
	"os"
	"revit/internal/db"

	"gopkg.in/yaml.v3"
)

type ContextConfigStore interface {
	GetUnusedContexts(ctx context.Context) ([]db.Context, error)
	AddContext(ctx context.Context, name string) error
	DeleteContextsById(ctx context.Context, id int64) error
	ExecTx(context.Context, func(db.Querier) error) error
}

type ContextsConfigMgr struct {
	Contexts map[string]struct{}
	store    ContextConfigStore
}

func MakeContextsMgrAndLoad(store ContextConfigStore, configPath string) (*ContextsConfigMgr, []forDirResult, error) {
	contextsMgr := &ContextsConfigMgr{store: store}
	results, err := forDir(configPath, func(file *os.File) error {
		err := contextsMgr.LoadContexts(file)
		if err != nil {
			return err
		}
		return nil
	})

	return contextsMgr, results, err
}

func (cm *ContextsConfigMgr) LoadContexts(f *os.File) error {
	dec := yaml.NewDecoder(f)
	var ctxs struct {
		Contexts []string `yaml:"contexts" json:"contexts"`
	}

	err := dec.Decode(&ctxs)
	if err != nil {
		return err
	}

	contextMap := make(map[string]struct{})
	for _, c := range ctxs.Contexts {
		_, found := contextMap[c]
		if found {
			//TODO: show error about repeating contexts
		}
		contextMap[c] = struct{}{}
	}

	cm.Contexts = contextMap
	return nil
}

func (cm *ContextsConfigMgr) SyncContexts(ctx context.Context) error {
	unused, err := cm.store.GetUnusedContexts(ctx)
	if err != nil {
		return err
	}

	orphanContextsIds := []int64{}

	for _, v := range unused {
		if _, found := cm.Contexts[v.Name]; !found {
			orphanContextsIds = append(orphanContextsIds, v.ID)
		}
	}

	err = cm.store.ExecTx(ctx, func(q db.Querier) error {
		for _, id := range orphanContextsIds {
			err := q.DeleteContextsById(ctx, id)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
