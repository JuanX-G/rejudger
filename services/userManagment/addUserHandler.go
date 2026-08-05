package userManagmentService

import (
	"net/http"
	"revit/internal/db"
	"revit/internal/passwords"

	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserAddQuery struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	InternaId string `json:"internal_id"`
}

type UserManagmentResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

func (s *UserManagementService) addUserHandler() (func(http.ResponseWriter, *http.Request)) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)  {
		var query UserAddQuery
		if _, err := httpHelpers.RequestJsonToStruct(r, &query); err != nil {
			writeUserManagmentFail(w, http.StatusBadRequest, "invalid request")
			return
		}
		ctx, cancel := s.requestCtx(r)
		defer cancel()

		hash, err := passwords.MakeHash(query.Password)
		if err != nil {
			writeUserManagmentFail(w, http.StatusInternalServerError, "error occured while inserting the user record")
			return
		}
		err = s.store.Queries.InsertUser(ctx, db.InsertUserParams{Name: query.Name, Password: hash,
			InternalID: pgtype.Text{String: query.InternaId, Valid: query.InternaId != ""}, Email: pgtype.Text{String: query.Email, Valid: query.Email != ""}})
		if err != nil {
			writeUserManagmentFail(w, http.StatusInternalServerError, "error occured while inserting the user record")
			return
		}
		jsonHelpers.WriteJSON(w, http.StatusOK, "user created succesfully")
	})
}
