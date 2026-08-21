package testingservices

import (
	"revit/internal/permissions"
	"revit/services/auth"
)

const DEFAULT_USERNAME = "user"
const DEFAULT_USERID = 1
const DEFAULT_EXPIRY = 120

var DEFAULT_PERMISSION_SET = permissions.PermissionSet{
	Context:     DEFAULT_APPCONTEXT,
	Permissions: DEFAULT_PERMS,
}

func MakeAuthMgr(perms permissions.PermissionSet) (string, *auth.AuthManager, error) {
	mgr := auth.NewAuthManager()
	token, err := mgr.NewSession(DEFAULT_EXPIRY, DEFAULT_USERID, DEFAULT_USERNAME, DEFAULT_PERMISSION_SET)
	return token, mgr, err
}
