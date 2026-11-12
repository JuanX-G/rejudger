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
const DEFAULT_PERMS_ID2 = 2

const DEFAULT_PERMS_CONTEXT1 = "1"
const DEFAULT_PERMS_CONTEXT2 = "2"

var DEFAULT_PERMS_SLC = []db.Permission{
	db.Permission{ID: DEFAULT_PERMS_ID, Context: DEFAULT_PERMS_ID, Action: PermissionSubmit.String()},
	db.Permission{ID: DEFAULT_PERMS_ID2, Context: DEFAULT_PERMS_ID2, Action: PermissionPlusOne.String()},
}

var DEFAULT_PERMISSIONS_SETS = []PermissionSet{
	PermissionSet{Context: DEFAULT_PERMS_CONTEXT1, Permissions: map[ActionPermission]struct{}{PermissionSubmit: struct{}{}}},
	PermissionSet{Context: DEFAULT_PERMS_CONTEXT2, Permissions: map[ActionPermission]struct{}{PermissionPlusOne: struct{}{}}},
}

const DEFAULT_PERMISSION_SET_1_ACTION = PermissionSubmit

var DEFAULT_ACTIONS = []ActionPermission{PermissionSubmit, PermissionPlusOne}

const DEFAULT_ROLE_ID = 22
const DEFAULT_ROLE_NAME = "R"

type mockBasePermissionMgrStore struct {
	db.Querier

	PermInserted         []db.InsertPermissionParams
	PermDeleted          db.DeletePermissionParams
	PermDeletedByContext int64

	RoleInserted    string
	PermAddedToRole db.AddPermissionToRoleParams
}

func (ps *mockBasePermissionMgrStore) GetUserPermissions(_ context.Context, id int64) ([]db.Permission, error) {
	if id == 24 {
		return []db.Permission{}, fmt.Errorf("ERROR")
	}
	fmt.Printf("perm set given: %+v\n", DEFAULT_PERMS_SLC)
	return DEFAULT_PERMS_SLC, nil
}

func (ps *mockBasePermissionMgrStore) InsertPermission(_ context.Context, arg db.InsertPermissionParams) error {
	ps.PermInserted = append(ps.PermInserted, arg)
	return nil
}

func (ps *mockBasePermissionMgrStore) DeletePermission(_ context.Context, arg db.DeletePermissionParams) error {
	ps.PermDeleted = arg
	return nil
}

func (ps *mockBasePermissionMgrStore) DeletePermissionContext(_ context.Context, arg int64) error {
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

func (ps *mockBasePermissionMgrStore) GetContextById(_ context.Context, id int64) (db.Context, error) {
	if id == 1 {
		return db.Context{Name: DEFAULT_PERMS_CONTEXT1, ID: 1}, nil
	} else if id == 2 {
		return db.Context{Name: DEFAULT_PERMS_CONTEXT2, ID: 2}, nil
	}
	return db.Context{Name: "a", ID: 22}, nil
}

func (ps *mockBasePermissionMgrStore) GetContextByName(context.Context, string) (db.Context, error) {
	return db.Context{Name: "1", ID: 1}, nil
}

func (ps *mockBasePermissionMgrStore) ExecTx(_ context.Context, fn func(db.Querier) error) error {
	return fn(ps)
}
