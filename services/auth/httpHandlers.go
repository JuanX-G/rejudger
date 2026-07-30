package auth

import (
	"net/http"
	"revit/internal/permissions"
)

func (a *AuthManager) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if err := a.IsValid(token); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}

func (a *AuthManager) HasPermissionMiddleware(next http.HandlerFunc, context string, required permissions.ActionPermission) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if err := a.HasPermission(token, context, required); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}
