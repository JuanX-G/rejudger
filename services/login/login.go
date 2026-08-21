package login

import (
	"context"
	"net/http"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/internal/passwords"
	"revit/internal/permissions"
	"revit/services/auth"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type LoginManager struct {
	authService    *auth.AuthManager
	store          loginStore
	permissionsMgr permissions.PermissionsManager
}

func NewLoginManager(ctx context.Context, queries loginStore, authService *auth.AuthManager, permissionMgr permissions.PermissionsManager) (*LoginManager, error) {
	return &LoginManager{authService: authService, store: queries, permissionsMgr: permissionMgr}, nil
}

func (l *LoginManager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /session/login", l.LoginHandler())
	mux.HandleFunc("POST /session/logout", l.LogoutHandler())
}

type LoginQuery struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	UserName string `json:"user_name"`
}

type LoginResponse struct {
	Success bool   `json:"sucess"`
	Token   string `json:"token"`
}

func (l *LoginManager) LoginHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		var body LoginQuery
		_, err := httpHelpers.RequestJsonToStruct(r, &body)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, LoginResponse{Success: false})
			return
		}
		var pass string
		var id int
		var userName string
		if body.Email != "" {
			text := pgtype.Text{String: body.Email, Valid: true}
			data, err := l.store.GetBasicInfoByEmail(ctx, text)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Success: false})
				return
			}
			pass = data.Password
			id = int(data.ID)
			userName = data.Name
		} else if body.UserName != "" {
			data, err := l.store.GetBasicInfoByUserName(ctx, body.UserName)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Success: false})
				return
			}
			pass = data.Password
			id = int(data.ID)
			userName = data.Name
		} else {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, LoginResponse{Success: false})
			return
		}

		ok, err := passwords.VerifyPassword(pass, body.Password)
		if err != nil || !ok {
			jsonHelpers.WriteJSON(w, http.StatusUnauthorized, LoginResponse{Success: false})
			return
		}

		perms, err := l.permissionsMgr.GetUserPermissions(id)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Success: false})
			return
		}

		token, err := l.authService.NewSession(180, id, userName, permissions.PermissionSet{})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Success: false})
			return
		}
		for _, p := range perms {
			err := l.authService.UpdateSessionPerms(token, p)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Success: false})
			}
		}

		jsonHelpers.WriteJSON(w, http.StatusOK, LoginResponse{Success: true, Token: token})
	})
}
