package permissions

import (
	"slices"
	"testing"
)

// Count unique conexts in the PermissionSet slice
func countContexts(in []PermissionSet) int {
	seen := make(map[string]struct{})
	for _, set := range in {
		if _, ok := seen[set.Context]; !ok {
			seen[set.Context] = struct{}{}
		}
	}
	return len(seen)
}

func setUpPermissionMgr(t testing.TB) (PermissionsManager, *mockBasePermissionMgrStore) {
	mock := &mockBasePermissionMgrStore{}
	mgr, err := NewPermissionManager(t.Context(), mock)
	if err != nil {
		t.Fatalf("creating a new permission manager failed, error: %s", err)
	}
	return mgr, mock
}

// exists if the given permission set does not have at least one action from the ActionPermission slice.
func permissionSetExpectOne(t testing.TB, expected []ActionPermission, set PermissionSet) {
	found := false
	for _, p := range expected {
		_, ok := set.Permissions[p]
		if ok {
			found = true
		}
	}
	if !found {
		t.Fatalf("error: expected to find at least one of the follwing: %+v within the permissionsSets's map: %+v.", expected, set.Permissions)
	}
}

func TestGetUserPermissions(t *testing.T) {
	mgr, _ := setUpPermissionMgr(t)

	permSets, err := mgr.GetUserPermissions(t.Context(), 42)
	if err != nil {
		t.Fatalf("unexpected error returned from calling GetUserPermissions, found err: %s, expected err == nil", err)
	}

	expectedContexts := []string{DEFAULT_PERMS_CONTEXT1, DEFAULT_PERMS_CONTEXT2}
	seenContexts := make([]string, 0, 2)

	var usedAct ActionPermission
	for i, set := range permSets {
		_, ok := set.Permissions[usedAct]
		if ok {
			t.Fatalf("expected")
		}
		if i > 1 {
			t.Fatalf("expected two different contexts found: %d", countContexts(permSets))
		}
		seenContexts = append(seenContexts, set.Context)
		permissionSetExpectOne(t, DEFAULT_ACTIONS, set)
	}
	slices.Sort(expectedContexts)
	slices.Sort(seenContexts)
	if !slices.Equal(expectedContexts, seenContexts) {
		t.Fatalf("expected contexts: %+v, found: %+v", expectedContexts, seenContexts)
	}

	_, err = mgr.GetUserPermissions(t.Context(), 24) // causes the store to error
	if err == nil {
		t.Fatalf("get user permission recived an error from the data base but returned err == nil to the test.")
	}
}

func TestInsertPermission(t *testing.T) {
	mgr, store := setUpPermissionMgr(t)

	mgr.InsertPermissionSet(t.Context(), DEFAULT_PERMISSIONS_SETS[0])
	if len(store.PermInserted) != 1 {
		t.Fatalf("inserted one permission, found: %d in the store", len(store.PermInserted))
	}
	if store.PermInserted[0].Action != DEFAULT_PERMISSION_SET_1_ACTION.String() {
		t.Fatalf("found permission: %s in the store, expected: %s", store.PermInserted[0].Action, DEFAULT_PERMISSION_SET_1_ACTION)
	}
}
