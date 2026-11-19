package auth

import (
	"net/http"
	"revit/internal/jsonHelpers"
	"revit/internal/permissions"
)

// EndpointGuard owns a AuthManager interface and allows generating dependency
// injected middleware.
type EndpointGuard struct {
	authMgr AuthManager
}

func NewEndpointGuard(authMgr AuthManager) *EndpointGuard {
	return &EndpointGuard{authMgr: authMgr}
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

		appContext := r.Header.Get("X-App-Context")
		if appContext == "" {
			http.Error(w, "missing app context", http.StatusUnauthorized)
			return
		}

		ctx := ContextWithToken(r.Context(), token)
		r = r.WithContext(ctx)
		for perm := range perms.Permissions {
			if err := eg.authMgr.HasPermission(token, appContext, perm); err != nil {
				jsonHelpers.WriteJSON(w, http.StatusForbidden, map[string]string{
					"error": err.Error(),
				})
				return
			}

		}
		ctx = ContextWithAppCtx(ctx, appContext)

		authorId, err := eg.authMgr.GetUserId(token)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to authenticate",
			})
			return
		}

		ctx = ContextWithId(ctx, authorId)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
