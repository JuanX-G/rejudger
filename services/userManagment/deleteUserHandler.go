package userManagmentService

import (
	"net/http"
	"revit/internal/db"
	"revit/internal/httpHelpers"
)

func (s *UserManagementService) deleteUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query UserActionQuery
		defer r.Body.Close()
		if _, err := httpHelpers.RequestJsonToStruct(r, &query); err != nil {
			writeUserManagmentFail(w, http.StatusBadRequest, "invalid request")
			return
		}
		ctx, cancel := s.requestCtx(r)
		defer cancel()

		queryFn := func(queries *db.Queries) error {
			user, err := getUserDataFromQuery(ctx, w, s.store.GetQueries(), query)
			if err != nil {
				return err
			}
			err = s.store.GetQueries().DeleteUser(ctx, user.ID)
			if err != nil {
				return err
			}
			return nil
		}

		err := s.store.ExecTx(ctx, queryFn)
		if err != nil {
			writeUserManagmentFail(w, http.StatusInternalServerError, "error in the db")
			return
		}
	}
}
