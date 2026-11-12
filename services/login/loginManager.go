// Login handles login and logout. It inserts sessions and then returns tokens
// to clients.
package login

import (
	"context"
	"net/http"
	"revit/internal/auth"
	"revit/internal/httpHelpers"
	"revit/internal/permissions"
)

type LoginManager struct {
	authService    auth.AuthManager
	store          loginStore
	permissionsMgr permissions.PermissionsManager
}

func NewLoginManager(ctx context.Context, queries loginStore, authService auth.AuthManager, permissionMgr permissions.PermissionsManager) *LoginManager {
	return &LoginManager{authService: authService, store: queries, permissionsMgr: permissionMgr}
}

func (lm *LoginManager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /session/login", httpHelpers.Adapt(lm.Login))
	mux.HandleFunc("GET /session/logout", lm.LogoutHandler())
}
