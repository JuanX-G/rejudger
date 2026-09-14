package auth

import (
	"context"
	"net/http"
	"revit/internal/jsonHelpers"
	"revit/internal/permissions"
)

// EndpointGuard owns a AuthManager interface and allows generating dependency
// injected middleware.
type EndpointGuard struct {
	Context string // Context of permissions required by the guard
	authMgr AuthManager
}

func NewEndpointGuard(authMgr AuthManager, context string) *EndpointGuard {
	return &EndpointGuard{authMgr: authMgr, Context: context}
}

const CONTEXT_ID_KEY = "ContextIDKey"

func contextWithId(ctx context.Context, id int64) context.Context {
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

// The Lock method returns an http handlerFunc. It will be wrapped in
// middleware that enforces, the session for the request's token, has
// permissions matching the 'perms' permissionSet.
func (eg *EndpointGuard) Lock(next http.HandlerFunc, perms permissions.PermissionSet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			http.Error(w, "missing authentication token", http.StatusUnauthorized)
			return
		}

		ctx := ContextWithToken(r.Context(), token)
		r = r.WithContext(ctx)
		for perm := range perms.Permissions {
			if err := eg.authMgr.HasPermission(token, eg.Context, perm); err != nil {
				jsonHelpers.WriteJSON(w, http.StatusForbidden, map[string]string{
					"error": err.Error(),
				})
				return
			}

		}

		authorId, err := eg.authMgr.GetUserId(token)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to authenticate",
			})
			return
		}

		ctx = contextWithId(ctx, authorId)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
