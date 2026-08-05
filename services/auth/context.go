package auth

import (
	"context"
	"errors"
	"fmt"
	"revit/internal/permissions"
)

const SessionTokenKey string = "sessionToken"

func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, SessionTokenKey, token)
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(SessionTokenKey).(string)
	return token, ok
}

var ErrUnauthorized = errors.New("unauthorized: missing session token")

type AccessGuard struct {
	authMgr *AuthManager
}

func NewAccessGuard(authMgr *AuthManager) *AccessGuard {
	return &AccessGuard{authMgr: authMgr}
}


func (g *AccessGuard) Require(ctx context.Context, appContext string, required ...permissions.ActionPermission) error {
	token, ok := TokenFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	if err := g.authMgr.HasAllPermissions(token, appContext, required...); err != nil {
		return fmt.Errorf("permission denied [%s]: %w", appContext, err)
	}

	return nil
}
