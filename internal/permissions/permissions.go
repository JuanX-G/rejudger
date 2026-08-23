package permissions

import (
	"strings"
)

type ActionPermission int

//go:generate stringer -type=ActionPermission -trimprefix=Permission
const (
	PermissionUnknownAction ActionPermission = iota
	PermissionPlusOne
	PermissionSoftVeto
	PermissionHardVeto
	PermissionView
	PermissionSubmit
	PermissionWrite
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

type PermissionSet struct {
	Context     string
	Permissions map[ActionPermission]struct{}
}

//go:generate stringer -type=PermissionConfigErrorType -trimprefix=Permission
type PermissionConfigErrorType int

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

func HasPermission(has PermissionSet, req ActionPermission) bool {
	if _, found := has.Permissions[req]; found {
		return true
	} else {
		return false
	}
}
