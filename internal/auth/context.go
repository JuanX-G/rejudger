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
	authMgr AuthManager
}

func NewAccessGuard(authMgr AuthManager) *AccessGuard {
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

const CONTEXT_APPCONTEXT_KEY = "ContextAppContextKey"

func ContextWithAppCtx(ctx context.Context, appContext string) context.Context {
	return context.WithValue(ctx, CONTEXT_APPCONTEXT_KEY, appContext)
}

func GetAppContextFromCtx(ctx context.Context) (string, bool) {
	appContext, ok := ctx.Value(CONTEXT_APPCONTEXT_KEY).(string)
	if !ok {
		return "", false
	} else {
		return appContext, true
	}
}

const CONTEXT_ID_KEY = "ContextIDKey"

func ContextWithId(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, CONTEXT_ID_KEY, id)
}

func GetIdFromContext(ctx context.Context) (int64, bool) {
	id := ctx.Value(CONTEXT_ID_KEY)
	if id == nil {
		return -1, false
	} else {
		idint, ok := id.(int64)
		if !ok {
			return -1, false
		}
		return idint, true
	}
}
