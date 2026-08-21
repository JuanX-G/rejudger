package testingservices

import (
	"revit/internal/permissions"
	"revit/services/auth"
)

func MakeAuthMgr(perms permissions.PermissionSet) (string, *auth.AuthManager, error) {
	mgr := auth.NewAuthManager()
	token, err := mgr.NewSession(DEFAULT_EXPIRY, DEFAULT_USERID, DEFAULT_USERNAME, DEFAULT_PERMISSION_SET)
	return token, mgr, err
}
