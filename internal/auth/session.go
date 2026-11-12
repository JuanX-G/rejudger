package auth

import (
	"revit/internal/permissions"
	"time"
)

type Session struct {
	Expiry      time.Time // Time after which the sessions becomes invalid
	UserId      int
	UserName    string
	Permissions map[string]permissions.PermissionSet //map a context to the permissiosn the user has in it
}

// expiry duration amount of minutes
func newSession(expiryDur int, userId int, userName string, perms permissions.PermissionSet) Session {
	return Session{Expiry: time.Now().Add(time.Duration(expiryDur * 60 * 1000000000)),
		UserId: userId, UserName: userName, Permissions: map[string]permissions.PermissionSet{perms.Context: perms},
	}
}

func UpdateSessionPerms(s *Session, perms permissions.PermissionSet) {
	s.Permissions[perms.Context] = perms
}
