package testingservices

import (
	"context"
	"revit/internal/db"
	"revit/internal/permissions"
)

type MockPermissionsManager struct {
	Perms            permissions.PermissionSet
	Role             db.Role
	RolePermsissions MockRolePermissions
}

type MockRolePermissions struct {
	RoleId       int32
	PermissionId int32
}

const DEFAULT_ROLE_ID = 42
const DEFAULT_ROLE_NAME = "roler"

const DEFAULT_PERMISSION_ID = 142

var DEFAULT_ROLE = db.Role{ID: DEFAULT_ROLE_ID, Name: DEFAULT_ROLE_NAME}

func (p *MockPermissionsManager) GetUserPermissions(_ context.Context, userId int) ([]permissions.PermissionSet, error) {
	return []permissions.PermissionSet{p.Perms}, nil
}

func (p *MockPermissionsManager) InsertPermission(ctx context.Context, context string, action permissions.ActionPermission) error {
	if p.Perms.Permissions != nil && p.Perms.Context == context {
		p.Perms.Permissions[action] = struct{}{}
	}
	return nil
}

func (p *MockPermissionsManager) InsertPermissionSet(ctx context.Context, set permissions.PermissionSet) error {
	p.Perms = set
	return nil
}

func (p *MockPermissionsManager) DeletePermission(ctx context.Context, context string, action permissions.ActionPermission) error {
	if p.Perms.Permissions != nil && p.Perms.Context == context {
		delete(p.Perms.Permissions, action)
	}
	return nil
}

func (p *MockPermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	if p.Perms.Permissions != nil && p.Perms.Context == context {
		p.Perms = permissions.PermissionSet{}
	}
	return nil
}

func (p *MockPermissionsManager) InsertRole(ctx context.Context, name string) error {
	p.Role.Name = name
	p.Role.ID = 1
	return nil
}

func (p *MockPermissionsManager) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	if p.Role.Name != "" {
		return p.Role, nil
	}
	return DEFAULT_ROLE, nil
}

func (p *MockPermissionsManager) AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) error {
	p.RolePermsissions.PermissionId = permissionID
	p.RolePermsissions.RoleId = roleID
	return nil
}

func (p *MockPermissionsManager) GetPermissionId(ctx context.Context, context string, action permissions.ActionPermission) (int, error) {
	if p.Perms.Permissions != nil {
		return 1, nil
	}
	return DEFAULT_PERMISSION_ID, nil
}

func (p *MockPermissionsManager) GetExecTx() func(context.Context, func(*db.Queries) error) error {
	return nil
}

func (p *MockPermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms permissions.PermissionSet) error {
	p.InsertRole(ctx, roleName)
	p.InsertPermissionSet(ctx, perms)
	p.AddPermissionToRole(ctx, 1, p.Role.ID)
	return nil
}
