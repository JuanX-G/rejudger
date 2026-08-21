package testingservices

import (
	"context"
	"revit/internal/db"
	"revit/internal/permissions"
)

type MockPermissionsManager struct{}

const DEFAULT_ROLE_ID = 42
const DEFAULT_ROLE_NAME = "roler"

const DEFAULT_PERMISSION_ID = 142

var DEFAULT_ROLE = db.Role{ID: DEFAULT_ROLE_ID, Name: DEFAULT_ROLE_NAME}

func (p *MockPermissionsManager) GetUserPermissions(userId int) ([]permissions.PermissionSet, error) {
	return []permissions.PermissionSet{DEFAULT_PERMISSION_SET}, nil
}

func (p *MockPermissionsManager) InsertPermission(ctx context.Context, context string, action permissions.ActionPermission) error {
	return nil
}

func (p *MockPermissionsManager) InsertPermissionSet(ctx context.Context, set permissions.PermissionSet) error {
	return nil
}

func (p *MockPermissionsManager) DeletePermission(ctx context.Context, context string, action permissions.ActionPermission) error {
	return nil
}

func (p *MockPermissionsManager) DeletePermissionContext(ctx context.Context, context string) error {
	return nil
}

func (p *MockPermissionsManager) InsertRole(ctx context.Context, name string) error {
	return nil
}

func (p *MockPermissionsManager) GetRoleByName(ctx context.Context, name string) (db.Role, error) {
	return DEFAULT_ROLE, nil
}

func (p *MockPermissionsManager) AddPermissionToRole(ctx context.Context, permissionID int32, roleID int32) error {
	return nil
}

func (p *MockPermissionsManager) GetPermissionId(ctx context.Context, context string, action permissions.ActionPermission) (int, error) {
	return DEFAULT_PERMISSION_ID, nil
}

func (p *MockPermissionsManager) GetExecTx() func(context.Context, func(*db.Queries) error) error {
	return nil
}

func (p *MockPermissionsManager) AddRoleWithPermissions(ctx context.Context, roleName string, perms permissions.PermissionSet) error {
	return nil
}
