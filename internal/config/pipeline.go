package config

import (
)

type Pipeline struct {
	Name string `yaml:"name"`
	Stages []PipelineStage `yaml:"stages"`
}

func LoadPipelineConfig(fileName string) (*Pipeline, error) {
	pipeline := Pipeline{}
	err:= decodeFile(fileName, &pipeline)
	if err != nil {
		return nil, ConfigError{errType: ConfigErrorOpenFile}
	}
	for i := 0; i < len(pipeline.Stages); i++ {
		pipeline.Stages[i].Deadline.FrequencyStr = pipeline.Stages[i].DeadlineStr
		if err := pipeline.Stages[i].Deadline.ParseDeadline(); err != nil {
			return nil, err
		}
	}
	return &pipeline, nil
}

type PipelineStage struct {
	Name string `yaml:"name"`
	PlusOnesRequired int `yaml:"plus_ones_required"`
	Blind bool `yaml:"blind"`
	SoftVeto bool `yaml:"soft_veto"`
	HasDeadline bool `yaml:"has_deadline"`
	DeadlineStr string `yaml:"deadline"`
	Deadline StageDeadline
}
