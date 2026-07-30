package login

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	db "revit/internal/db"
	"revit/internal/passwords"
	"revit/internal/permissions"
	"revit/internal/jsonHelpers"
	"revit/services/auth"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginManager struct {
	authService *auth.AuthManager
	queries *db.Queries
	permissionsMgr *permissions.PermissionsManager
}

func NewLoginManager(ctx context.Context, pool *pgxpool.Pool, authService *auth.AuthManager, permissionMgr *permissions.PermissionsManager) (*LoginManager, error) {
	return &LoginManager{authService: authService, queries: db.New(pool), permissionsMgr: permissionMgr}, nil
}

type LoginQuery struct {
	Password string `json:"password"`
	Email string `json:"email"`
	UserName string `json:"user_name"`
}

type LoginResponse struct {
	Succes bool `json:"sucess"`
	Token string `json:"token"`
}

func (l *LoginManager) LoginHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 10)
		defer cancel()
		bd := r.Body
		bodyB, err := io.ReadAll(bd)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, LoginResponse{Succes: false})
			return
		}
		var body LoginQuery
		err = json.Unmarshal(bodyB, &body)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, LoginResponse{Succes: false})
			return
		}
		var pass string
		var id int
		var userName string
		if body.Email != "" {
			text := pgtype.Text{String: body.Email, Valid: true}
			data, err := l.queries.GetBasicInfoByEmail(ctx, text)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: false})
				return
			}
			pass = data.Password
			id = int(data.ID)
			userName = data.Name
		} else if body.UserName != "" {
			data, err := l.queries.GetBasicInfoByUserName(ctx, body.UserName)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: false})
				return
			}
			pass = data.Password
			id = int(data.ID)
			userName = data.Name
		} else {
				jsonHelpers.WriteJSON(w, http.StatusBadRequest, LoginResponse{Succes: false})
				return
		}

		ok, err := passwords.VerifyPassword(pass, body.Password)
		if err != nil || !ok {
			jsonHelpers.WriteJSON(w, http.StatusUnauthorized, LoginResponse{Succes: false})
			return
		}

		perms, err := l.permissionsMgr.GetUserPermissions(id)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: false})
			return
		}

		token, err := l.authService.NewSession(180, id, userName, permissions.PermissionSet{})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: false})
			return
		}
		for _, p := range perms {
			err := l.authService.UpdateSessionPerms(token, p)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: false})
			}
		}

		jsonHelpers.WriteJSON(w, http.StatusInternalServerError, LoginResponse{Succes: true, Token: token})
	})
}
