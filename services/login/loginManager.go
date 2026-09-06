// Login handles login and logout. It inserts sessions and then returns tokens
// to clients.
package login

import (
	"context"
	"net/http"
	"revit/internal/permissions"
	"revit/services/auth"
)

type LoginManager struct {
	authService    auth.AuthManager
	store          loginStore
	permissionsMgr permissions.PermissionsManager
}

func NewLoginManager(ctx context.Context, queries loginStore, authService auth.AuthManager, permissionMgr permissions.PermissionsManager) (*LoginManager, error) {
	return &LoginManager{authService: authService, store: queries, permissionsMgr: permissionMgr}, nil
}

func (l *LoginManager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /session/login", l.LoginHandler())
	mux.HandleFunc("GET /session/logout", l.LogoutHandler())
}
