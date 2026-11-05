package config

import (
	"fmt"
	"os"
	"revit/internal/permissions"
	"strings"

	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name  string             `yaml:"name"`
	Perms []PermissionConfig `yaml:"permissions"`
}

func LoadRolesConfigFromFile(file *os.File) (RoleConfig, error) {
	dec := yaml.NewDecoder(file)
	role := RoleConfig{}
	err := dec.Decode(&role)
	if err != nil {
		return RoleConfig{}, err
	}

	for i := range role.Perms {
		if err := role.Perms[i].ParsePermissionConfig(); err != nil {
			return RoleConfig{}, err
		}
	}

	return role, nil
}

// Config for a permission.
//
// # Permission String Format
// context :: action-allowed
type PermissionConfig struct {
	PermissionStr string                       `yaml:"permission"` // Permissions as the original permission string
	Action        permissions.ActionPermission `yaml:"-"`
	Context       string                       `yaml:"-"`
}

func (p *PermissionConfig) ParsePermissionConfig() error {
	parts := strings.Split(p.PermissionStr, "::")
	if len(parts) != 2 {
		return ConfigError{errType: ConfigErrorInvaidFormatting, msg: fmt.Sprintf(`invalid permission string format,
			found string: %s, with: %d part when split by '::'`, p.PermissionStr, len(parts))}
	}

	p.Context = strings.TrimSpace(parts[0])
	action, err := permissions.ParseActionPermission(strings.TrimSpace(parts[1]))
	if err != nil {
		return err
	}
	p.Action = action
	return nil
}
