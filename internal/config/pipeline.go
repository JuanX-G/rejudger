package config

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Pipeline struct {
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description" json:"description"`
	Stages      []PipelineStage `yaml:"stages" json:"stages"`
}

func LoadPipelineConfig(fileName string) (*Pipeline, error) {
	pipeline := Pipeline{}
	err := decodeFile(fileName, &pipeline)
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

func (p *Pipeline) JSON() ([]byte, error) {
	return json.Marshal(p)
}

type PipelineStage struct {
	Name             string `yaml:"name" json:"name"`
	PlusOnesRequired int16  `yaml:"plus_ones_required" json:"plus_ones_required"`
	Blind            bool   `yaml:"blind" json:"blind"`
	SoftVeto         bool   `yaml:"soft_veto" json:"soft_veto"`
	HasDeadline      bool   `yaml:"has_deadline" json:"has_deadline"`
	DeadlineStr      string `yaml:"deadline" json:"deadline"`

	Deadline StageDeadline `json:"-"`
}

func boolToString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func (p *PipelineStage) Stringify() string {
	sb := strings.Builder{}
	sb.WriteString(p.Name)
	sb.WriteString(strconv.Itoa(int(p.PlusOnesRequired)))
	sb.WriteString(boolToString(p.Blind))
	sb.WriteString(boolToString(p.SoftVeto))
	sb.WriteString(boolToString(p.HasDeadline))
	sb.WriteString(p.DeadlineStr)
	return sb.String()
}

func (p *PipelineStage) JSON() ([]byte, error) {
	return json.Marshal(p)
}
