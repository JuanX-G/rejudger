package config

type BaseConfig struct {
	InstanceName          string `yaml:"name"`
	RemoteUserManagment   string `yaml:"remote_user_managment"`
	RemoteUserManagmentOn bool
	PipelieConfigPath     string `yaml:"pipeline_config_path"`
	RoleConfigPath        string `yaml:"role_config_path"`
	UserConfigPath        string `yaml:"user_config_path"`
}

// Load the basic config that lists the locations of other configs and
// the ability to remotely manage users.
func LoadBaseConfig(fileName string) (*BaseConfig, error) {
	config := BaseConfig{}
	err := decodeFile(fileName, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
