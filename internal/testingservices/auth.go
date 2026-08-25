package testingservices

import (
	"revit/internal/permissions"
	"revit/services/auth"
	"time"
)

func MakeAuthMgr(perms permissions.PermissionSet) (string, *MockAuthManager, error) {
	mgr := &MockAuthManager{}
	token, err := mgr.NewSession(DEFAULT_EXPIRY, DEFAULT_USERID, DEFAULT_USERNAME, DEFAULT_PERMISSION_SET)
	return token, mgr, err
}

type MockAuthManager struct {
	SavedSession auth.Session
	full         bool
	ReturnErrors bool
}

func (am *MockAuthManager) DeleteToken(token string) error {
	am.full = false
	am.SavedSession = auth.Session{}
	return nil
}

func (am *MockAuthManager) NewSession(expiryDur int, userId int, userName string, perms permissions.PermissionSet) (string, error) {
	am.SavedSession.UserId = userId
	am.SavedSession.UserName = userName
	am.SavedSession.Permissions = map[string]permissions.PermissionSet{
		perms.Context: perms,
	}
	am.SavedSession.Expiry = time.Time{}
	am.full = true
	return DEFAULT_SESSION_TOKEN, nil
}

func (am *MockAuthManager) UpdateSessionPerms(token string, perms permissions.PermissionSet) error {
	if !am.full {
		return auth.ErrNoSession
	}
	am.SavedSession.Permissions[perms.Context] = perms
	return nil
}

func (am *MockAuthManager) HasPermission(token string, context string, required permissions.ActionPermission) error {
	if !am.full {
		return auth.ErrNoSession
	}
	perm, ok := am.SavedSession.Permissions[context]
	if !ok {
		return auth.ErrorNoSuchSession
	}
	if !permissions.HasPermission(perm, required) {
		return auth.ErrForbidden
	} else {
		return nil
	}
}

func (am *MockAuthManager) IsValid(token string) error {
	if am.ReturnErrors {
		return auth.ErrExpired
	}
	return nil
}

func (am *MockAuthManager) GetUserId(token string) (int64, error) {
	if am.ReturnErrors {
		return -1, auth.ErrorNoSuchSession
	}
	if !am.full {
		return -1, auth.ErrNoSession
	}
	return int64(am.SavedSession.UserId), nil
}

func (am *MockAuthManager) HasAllPermissions(token string, context string, required ...permissions.ActionPermission) error {
	if am.ReturnErrors {
		return auth.ErrForbidden
	}

	perms, ok := am.SavedSession.Permissions[context]
	if !ok {
		return auth.ErrForbidden
	}

	for _, req := range required {
		if !permissions.HasPermission(perms, req) {
			return auth.ErrForbidden
		}
	}

	return nil
}
