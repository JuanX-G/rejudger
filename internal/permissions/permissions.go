package permissions

import (
	"strings"
)

type ActionPermission int

//go:generate stringer -type=ActionPermission -trimprefix=Permission
const (
    PermissionPlusOne ActionPermission = iota
    PermissionSoftVeto
	PermissionHardVeto
	PermissionView
	PermissionSubmit
	PermissionManage
	PermissionUnknownAction
)

func ParseActionPermission(s string) (ActionPermission, error) {
	switch s {
	case "PlusOne":
		return PermissionPlusOne, nil
	case "SoftVeto":
		return PermissionSoftVeto, nil
	case "HardVeto":
		return PermissionHardVeto, nil
	case "View":
		return PermissionView, nil
	case "Submit":
		return PermissionSubmit, nil
	case "Manage":
		return PermissionManage, nil
	default:
		return PermissionUnknownAction, PermissionConfigError{errType: PermissionConfigErrorInvalidAction}
	}
}

type PermissionSet struct {
	Context string
	Permissions map[ActionPermission]struct{}
}

type PermissionConfigErrorType int
const (
	PermissionConfigErrorInvalidString PermissionConfigErrorType = iota
	PermissionConfigErrorInvalidPermissionSet
	PermissionConfigErrorInvalidAction
)

type PermissionConfigError struct {
	errType PermissionConfigErrorType
	customMessage string
}

func (p PermissionConfigError) Error() string {
	switch p.errType {
	case PermissionConfigErrorInvalidString:
		return "inavlid string format; possibly missing ':'"
	case PermissionConfigErrorInvalidPermissionSet:
		return "inavlid format for the set of permissions"
	case PermissionConfigErrorInvalidAction:
		return "invalid action name"
	default:
		return "unknown error with permission string"
	}
}

func NewPermissionSetFromString(s string) (PermissionSet, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return PermissionSet{}, PermissionConfigError{errType: PermissionConfigErrorInvalidString}
	}
	retSet := PermissionSet{Context: parts[0]}
	perms := strings.ReplaceAll(parts[1], " ", "")
	permsParts := strings.SplitSeq(perms, ",")
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
