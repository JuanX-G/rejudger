package login

import (
	"errors"
	"net/http"
	"revit/internal/jsonHelpers"
	"revit/services/auth"
)

type logoutResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

func (l *LoginManager) LogoutHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, logoutResponse{Success: false, Message: "token not included"})
			return
		}
		if err := l.authService.DeleteToken(token); err != nil {
			if ok := errors.Is(err, auth.ErrorNoSuchSession); ok {
				jsonHelpers.WriteJSON(w, http.StatusBadRequest, logoutResponse{Success: false, Message: "token not found"})
				return
			} else {
				jsonHelpers.WriteJSON(w, http.StatusInternalServerError, logoutResponse{Success: false, Message: "internal server error"})
				return
			}
		}
		jsonHelpers.WriteJSON(w, http.StatusOK, logoutResponse{Success: true, Message: "logged out"})
		return
	})

}
