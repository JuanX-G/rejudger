package permissions

import (
	"strings"
)

//go:generate stringer -type=ActionPermission -trimprefix=Permission
type ActionPermission int

// Posssible permission for a single appContext.
const (
	PermissionUnknownAction ActionPermission = iota // Invalid state, arises when parsing external strings into permissions
	PermissionPlusOne                               // Permission to approve a submission
	PermissionSoftVeto                              // Permission to soft-veto a submission
	PermissionHardVeto                              // Permission to hard-veto a submission
	PermissionView                                  // Permission to view all submissions
	PermissionSubmit                                // Permission to make submissions
	PermissionWrite                                 // Permission to write to shared resources and configurations in the context
)

// Parse a permission name as string to a type.
func ParseActionPermission(s string) (ActionPermission, error) {
	switch s {
	case PermissionPlusOne.String():
		return PermissionPlusOne, nil
	case PermissionSoftVeto.String():
		return PermissionSoftVeto, nil
	case PermissionHardVeto.String():
		return PermissionHardVeto, nil
	case PermissionView.String():
		return PermissionView, nil
	case PermissionSubmit.String():
		return PermissionSubmit, nil
	case PermissionWrite.String():
		return PermissionWrite, nil
	default:
		return PermissionUnknownAction, PermissionConfigError{errType: PermissionConfigErrorInvalidAction}
	}
}

// Struct representing a set of permisions for a certain context in the app.
type PermissionSet struct {
	Context     string                        // Application context where such permissions hold
	Permissions map[ActionPermission]struct{} // The permisions held
}

//go:generate stringer -type=PermissionConfigErrorType -trimprefix=Permission
type PermissionConfigErrorType int

// Possible errors when parsing the string representation of permissions.
const (
	PermissionConfigErrorInvalidString PermissionConfigErrorType = iota
	PermissionConfigErrorInvalidPermissionSet
	PermissionConfigErrorInvalidAction
)

type PermissionConfigError struct {
	errType       PermissionConfigErrorType
	customMessage string
}

func (p PermissionConfigError) Error() string {
	switch p.errType {
	case PermissionConfigErrorInvalidString:
		return "inavlid string format; possibly missing '::'"
	case PermissionConfigErrorInvalidPermissionSet:
		return "inavlid format for the set of permissions"
	case PermissionConfigErrorInvalidAction:
		return "invalid action name"
	default:
		return "unknown error with permission string"
	}
}

// Make a new permission from a string. Format: <AppContext>::<Permission-list>.
// The Permission-list should be separated by ',' and can include spaces.
func NewPermissionSetFromString(s string) (PermissionSet, error) {
	parts := strings.Split(s, "::")
	if len(parts) != 2 {
		return PermissionSet{}, PermissionConfigError{errType: PermissionConfigErrorInvalidString}
	}
	retSet := PermissionSet{Context: parts[0]}
	perms := strings.ReplaceAll(parts[1], " ", "")
	permsParts := strings.SplitSeq(perms, ",")
	retSet.Permissions = make(map[ActionPermission]struct{})
	for p := range permsParts {
		action, err := ParseActionPermission(p)
		if err != nil {
			return PermissionSet{}, err
		}
		retSet.Permissions[action] = struct{}{}
	}
	return retSet, nil
}

// Checks if set has the permission to perform action `req`.
func HasPermission(set PermissionSet, req ActionPermission) bool {
	if _, found := set.Permissions[req]; found {
		return true
	} else {
		return false
	}
}
