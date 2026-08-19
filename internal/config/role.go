package config

import (
	"os"
	"revit/internal/permissions"
	"strings"

	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	name  string
	perms []PermissionConfig
}

func LoadRolesConfigFromFile(file *os.File) (RoleConfig, error) {
	dec := yaml.NewDecoder(file)
	role := RoleConfig{}
	err := dec.Decode(&role)
	if err != nil {
		return RoleConfig{}, err
	}
	for _, perm := range role.perms {
		if err := perm.ParsePermissionConfig(); err != nil {
			return RoleConfig{}, err
		}
	}
	return role, nil
}

type PermissionConfig struct {
	permissionStr string `yaml:"permission"`
	action        permissions.ActionPermission
	context       string
}

func (p *PermissionConfig) ParsePermissionConfig() error {
	parts := strings.Split(p.permissionStr, "::")
	if len(parts) != 2 {
		return ConfigError{}
	}
	p.context = strings.TrimSpace(parts[0])
	action, err := permissions.ParseActionPermission(strings.TrimSpace(parts[1]))
	if err != nil {
		return err
	}
	p.action = action
	return nil
}
