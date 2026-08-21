package testingservices

import "revit/internal/permissions"

const DEFAULT_APPCONTEXT = "context"

var DEFAULT_PERMS = map[permissions.ActionPermission]struct{}{
	permissions.PermissionSoftVeto: struct{}{},
	permissions.PermissionHardVeto: struct{}{},
	permissions.PermissionPlusOne:  struct{}{},
	permissions.PermissionSubmit:   struct{}{},
	permissions.PermissionView:     struct{}{},
	permissions.PermissionWrite:    struct{}{},
}
