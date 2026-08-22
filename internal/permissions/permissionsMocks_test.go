package permissions

import (
	"context"
	"fmt"
	"revit/internal/db"
)

var DEFAULT_PERMS = map[ActionPermission]struct{}{
	PermissionSoftVeto: struct{}{},
	PermissionHardVeto: struct{}{},
	PermissionPlusOne:  struct{}{},
	PermissionSubmit:   struct{}{},
	PermissionView:     struct{}{},
	PermissionWrite:    struct{}{},
}

const DEFAULT_PERMS_ID = 1
const DEFAULT_PERMS_ID2 = 12

const DEFAULT_PERMS_CONTEXT1 = "C1"
const DEFAULT_PERMS_CONTEXT2 = "C2"

var DEFAULT_PERMS_SLC = []db.Permission{
	db.Permission{ID: DEFAULT_PERMS_ID, Context: DEFAULT_PERMS_CONTEXT1, Action: PermissionSubmit.String()},
	db.Permission{ID: DEFAULT_PERMS_ID2, Context: DEFAULT_PERMS_CONTEXT2, Action: PermissionPlusOne.String()},
}

var DEFAULT_PERMISSIONS_SETS = []PermissionSet{
	PermissionSet{Context: DEFAULT_PERMS_CONTEXT1, Permissions: map[ActionPermission]struct{}{PermissionSubmit: struct{}{}}},
	PermissionSet{Context: DEFAULT_PERMS_CONTEXT2, Permissions: map[ActionPermission]struct{}{PermissionPlusOne: struct{}{}}},
}

var DEFAULT_ACTIONS = []ActionPermission{PermissionSubmit, PermissionPlusOne}

const DEFAULT_ROLE_ID = 22
const DEFAULT_ROLE_NAME = "R"

type mockBasePermissionMgrStore struct {
	PermInserted         db.InsertPermissionParams
	PermDeleted          db.DeletePermissionParams
	PermDeletedByContext string

	RoleInserted    string
	PermAddedToRole db.AddPermissionToRoleParams
}

func (ps *mockBasePermissionMgrStore) GetUserPermissions(_ context.Context, id int64) ([]db.Permission, error) {
	if id == 24 {
		return []db.Permission{}, fmt.Errorf("ERROR")
	}
	return DEFAULT_PERMS_SLC, nil
}

func (ps *mockBasePermissionMgrStore) InsertPermission(_ context.Context, arg db.InsertPermissionParams) error {
	ps.PermInserted = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) DeletePermission(_ context.Context, arg db.DeletePermissionParams) error {
	ps.PermDeleted = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) DeletePermissionContext(_ context.Context, arg string) error {
	ps.PermDeletedByContext = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) InsertRole(_ context.Context, arg string) error {
	ps.RoleInserted = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) GetRoleByName(_ context.Context, arg string) (db.Role, error) {
	return db.Role{ID: DEFAULT_ROLE_ID, Name: DEFAULT_ROLE_NAME}, nil
}

func (ps *mockBasePermissionMgrStore) AddPermissionToRole(_ context.Context, arg db.AddPermissionToRoleParams) error {
	ps.PermAddedToRole = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) GetPermissionId(context.Context, db.GetPermissionIdParams) (int32, error) {
	return DEFAULT_PERMS_ID, nil
}
