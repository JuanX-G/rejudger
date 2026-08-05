package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigMgr struct {
	Base *BaseConfig
	Pipelines []Pipeline
}

func NewConfigMgr(fileName string) (*ConfigMgr, error) {
	base, err := LoadBaseConfig(fileName)
	if err != nil {
		return nil, err
	}
	pipelines := make([]Pipeline, 0)
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
	return &ConfigMgr{Base: base, Pipelines: pipelines}, nil
}

