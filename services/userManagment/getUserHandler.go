package userManagmentService

import (
	"net/http"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
)

type GetUserResponse struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	InternaId string   `json:"internal_id"`
	Roles     []string `json:"roles"`
}

func (s *UserManagementService) getUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query UserActionQuery
		defer r.Body.Close()
		if _, err := httpHelpers.RequestJsonToStruct(r, &query); err != nil {
			writeUserManagmentFail(w, http.StatusBadRequest, "invalid request")
			return
		}
		ctx, cancel := s.requestCtx(r)
		defer cancel()

		var user db.User
		userRoles := []db.Role{}
		queryFn := func(queries db.Querier) error {
			user, err := getUserDataFromQuery(ctx, w, queries, query)
			if err != nil {
				return err
			}

			userRoles, err = queries.GetUserRoles(ctx, user.ID)
			if err != nil {
				return err
			}
			return nil
		}

		if err := s.store.ExecTx(ctx, queryFn); err != nil {
			writeUserManagmentFail(w, http.StatusInternalServerError, "db error")
			return
		}

		roleNames := []string{}
		for _, role := range userRoles {
			roleNames = append(roleNames, role.Name)
		}
		res := GetUserResponse{Name: user.Name, Roles: roleNames}
		if user.Email.Valid {
			res.Email = user.Email.String
		}
		if user.InternalID.Valid {
			res.InternaId = user.InternalID.String
		}
		jsonHelpers.WriteJSON(w, http.StatusOK, res)
	}
}
