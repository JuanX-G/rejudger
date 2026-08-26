package auth

import (
	"net/http"
	"revit/internal/jsonHelpers"
	"revit/internal/permissions"
)

type EndpointGuard struct {
	Context string
	authMgr AuthManager
}

func NewEndpointGuard(authMgr AuthManager, context string) *EndpointGuard {
	return &EndpointGuard{authMgr: authMgr, Context: context}
}

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
