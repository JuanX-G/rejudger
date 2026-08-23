package permissions

import (
	"errors"
	"testing"
)

func CheckParseInvalidActionError(err error, t testing.TB) {
	var permErr PermissionConfigError
	var ok bool
	if permErr, ok = errors.AsType[PermissionConfigError](err); !ok {
		t.Fatalf("expected error of type: %T, found: %T", PermissionConfigError{}, permErr)
	}

	if permErr.errType != PermissionConfigErrorInvalidAction {
		t.Fatalf("expected error type: %s, found: %s", PermissionConfigErrorInvalidAction, permErr.errType)
	}
}

func TestParsePermission(t *testing.T) {
	perm, err := ParseActionPermission("NOT_A_PERMISSION IN ANY WRODL!!")
	if perm != PermissionUnknownAction {
		t.Fatalf("expected permission: %s from parsing invalid string, found: %s", PermissionUnknownAction, perm)
	}

	CheckParseInvalidActionError(err, t)
}

func TestPermissionSetFromString(t *testing.T) {
	p, err := NewPermissionSetFromString("WRONG")
	if err == nil {
		t.Fatalf("invalid string passed to NewPermissionSetFromstring resulted in a nil error.")
	}
	if cfgErr, ok := errors.AsType[PermissionConfigError](err); ok {
		if cfgErr.errType != PermissionConfigErrorInvalidString {
			t.Fatalf("expected error type to be: 'PermissionConfigErrorInvalidString', found: %v", cfgErr.errType)
		}
	} else {
		t.Fatalf("expected error of type: %T to be returned, found: %T", PermissionConfigError{}, err)
	}
	if p.Permissions != nil {
		t.Fatalf("invalid string used to make a permission set, expected the returned set to have its permission map nil, found: %v", p.Permissions)
	}

	var permErr PermissionConfigError
	var ok bool
	if permErr, ok = errors.AsType[PermissionConfigError](err); !ok {
		t.Fatalf("expected error of type: %T, found: %T", PermissionConfigError{}, permErr)
	}

	if permErr.errType != PermissionConfigErrorInvalidString {
		t.Fatalf("expected error type: %s, found: %s", PermissionConfigErrorInvalidString, permErr.errType)
	}

	expected := map[ActionPermission]struct{}{
		PermissionWrite: struct{}{},
		PermissionView:  struct{}{},
	}
	p, err = NewPermissionSetFromString("C:: View, Write")
	if err != nil {
		t.Fatalf("provided a valid permission string, error: %s was returned still", err)
	}
	if p.Permissions == nil {
		t.Fatalf("provided a valid permission string but the permission map was found to be nil. PermissionSet returned: %s", p)
	}
	idx := 0
	for act, _ := range p.Permissions {
		if idx > 2 {
			t.Fatalf("permission string contained permission for two actions, third action was returned from parsing")
		}
		if _, ok := expected[act]; !ok {
			t.Fatalf("found an unexpected permssion: %s, in the returned permission set", act)
		}
		idx++
	}
	if idx != 2 {
		t.Fatalf("permission string contained permission for two actions, only: %d were found", idx)
	}
	if p.Context != "C" {
		t.Fatalf("expected conext to be: %s, found: %s", "C", p.Context)
	}

	p, err = NewPermissionSetFromString("C:: INVALIDPERMISSISOKAJIJIJIJ, Write")
	CheckParseInvalidActionError(err, t)

	if p.Permissions != nil {
		t.Fatalf("invalid string used to make a permission set, expected the returned set to have its permission map nil, found: %v", p.Permissions)
	}
}

func TestHasPermission(t *testing.T) {
	set := PermissionSet{Context: "C", Permissions: map[ActionPermission]struct{}{
		PermissionSubmit: struct{}{},
	}}
	has := PermissionSubmit
	if ok := HasPermission(set, has); !ok {
		t.Fatalf("Error HasPermission returned %t, on the set: %+v containing permission: %s", ok, set, has)
	}
	has = PermissionSoftVeto
	if ok := HasPermission(set, has); ok {
		t.Fatalf("Error HasPermission returned %t, on the set: %+v containing permission: %s", ok, set, has)
	}
}
