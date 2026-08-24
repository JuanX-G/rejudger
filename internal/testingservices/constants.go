package testingservices

import (
	"revit/internal/artifactmanager"
	"revit/internal/permissions"
)

const DEFAULT_APPCONTEXT = "context"
const DEFAULT_EXPIRY = 120
const DEFAULT_SESSION_TOKEN = "AB2137"

const DEFAULT_USERNAME = "user"
const DEFAULT_USERID = 1
const DEFAULT_PASSWORD = "pass"
const DEFAULT_EMAIL = "e@e.e"

const DEFAULT_OBJECT = "object"
const DEFAULT_OBJECT_HASH = "ohash"
const DEFAULT_OBJECT_NAME = "oname"

var DEFAULT_OBJECT_METADATA = artifactmanager.ArtifactMetadata{FileType: artifactmanager.ArtifactJPG, BaseHash: DEFAULT_OBJECT_HASH, Name: DEFAULT_OBJECT_NAME}
var DEFAULT_PERMISSION_SET = permissions.PermissionSet{
	Context:     DEFAULT_APPCONTEXT,
	Permissions: DEFAULT_PERMS,
}

var DEFAULT_PERMS = map[permissions.ActionPermission]struct{}{
	permissions.PermissionSoftVeto: struct{}{},
	permissions.PermissionHardVeto: struct{}{},
	permissions.PermissionPlusOne:  struct{}{},
	permissions.PermissionSubmit:   struct{}{},
	permissions.PermissionView:     struct{}{},
	permissions.PermissionWrite:    struct{}{},
}
