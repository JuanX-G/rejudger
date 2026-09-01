package auth

import (
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
		next(w, r)
	}
}
