package config

type BaseConfig struct {
	InstanceName string `yaml:"name"`
	RemoteUserManagment bool `yaml:"remote_user_managment"`
	PipelieConfigPath string `yaml:"pipeline_config_path"`
	RoleConfigPath string `yaml:"role_config_path"`
	UserConfigPath string `yaml:"user_config_path"`
}

func LoadBaseConfig(fileName string) (*BaseConfig, error) {
	config := BaseConfig{}
	err := decodeFile(fileName, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
