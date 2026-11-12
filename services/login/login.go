package login

import (
	"context"
	"net/http"
	"revit/internal/httpHelpers"
	"revit/internal/passwords"
	"revit/internal/permissions"

	"github.com/jackc/pgx/v5/pgtype"
)

// Query for login. If both email and username are provided then the email will
// be used. If none are provided an error is returned.
type LoginQuery struct {
	Password string `json:"password"`  // User password
	Email    string `json:"email"`     // User email
	UserName string `json:"user_name"` // User name
}

type LoginResponse struct {
	Success bool   `json:"sucess"` // True if logging succeeds
	Token   string `json:"token"`  // Token for the sessions created if login succeeds. Empty on failure
}

// Use [LoginQuery] to login with this handler. If either email or username
// is empty, then the non-empty one will be used; if both are empty a error
// will be sent back. If both are set, email will be used.
func (lm *LoginManager) Login(ctx context.Context, body LoginQuery) (LoginResponse, error) {
	var pass string
	var id int
	var userName string
	var err error

	if body.Email != "" {
		data, e := lm.store.GetBasicInfoByEmail(ctx, pgtype.Text{String: body.Email, Valid: true})
		err = e
		pass, id, userName = data.Password, int(data.ID), data.Name
	} else if body.UserName != "" {
		data, e := lm.store.GetBasicInfoByUserName(ctx, body.UserName)
		err = e
		pass, id, userName = data.Password, int(data.ID), data.Name
	} else {
		return LoginResponse{}, httpHelpers.NewError(http.StatusBadRequest, "username or email required")
	}

	if err != nil {
		return LoginResponse{}, httpHelpers.NewError(http.StatusUnauthorized, "invalid credentials")
	}

	ok, err := passwords.VerifyPassword(pass, body.Password)
	if err != nil || !ok {
		return LoginResponse{}, httpHelpers.NewError(http.StatusUnauthorized, "invalid credentials")
	}

	perms, err := lm.permissionsMgr.GetUserPermissions(ctx, id)
	if err != nil {
		return LoginResponse{}, err
	}

	token, err := lm.authService.NewSession(180, id, userName, permissions.PermissionSet{})
	if err != nil {
		return LoginResponse{}, err
	}

	for _, p := range perms {
		_ = lm.authService.UpdateSessionPerms(token, p)
	}

	return LoginResponse{Success: true, Token: token}, nil
}
