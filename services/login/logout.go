package login

import (
	"errors"
	"net/http"
	"revit/internal/jsonHelpers"
	"revit/services/auth"
)

type logoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Logut deletes a session with a set token. The token should be sent by the
// header: 'X-Auth-Token'; if the header is empty an error is returned.
func (l *LoginManager) LogoutHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, logoutResponse{Success: false, Message: "token not included"})
			return
		}
		defer r.Body.Close()
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
	})

}
