package userManagmentService

import (
	"net/http"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/internal/db"
)

type addToRoleQuery struct {
	User UserActionQuery `json:"user"`
	Roles []string `json:"roles"`
}

func (s *UserManagementService) addToRoleHandler() func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request)  {
		var query addToRoleQuery
		if _, err := httpHelpers.RequestJsonToStruct(r, &query); err != nil {
			writeUserManagmentFail(w, http.StatusBadRequest, "malformed request")
		}
		ctx, cancel := s.requestCtx(r)
		defer cancel()

		queryFn := func(q *db.Queries) error {
			roleIds := make([]int32, 0, len(query.Roles))
			for _, r := range query.Roles {
				role, err := q.GetRoleByName(ctx, r)
				if err != nil {
					return err
				}
				roleIds = append(roleIds, role.ID)
			}
			user, err := getUserDataFromQuery(ctx, w, q, query.User)
			if err != nil {
				return err
			}
			userId := user.ID
			for _, roleId := range roleIds {
				err := q.AddUserToRole(ctx, db.AddUserToRoleParams{RoleID: roleId, UserID: userId})
				if err != nil {
					return err
				}
			}
			return nil
		}
		if err := s.store.ExecTx(ctx, queryFn); err != nil {
			writeUserManagmentFail(w, http.StatusInternalServerError, "internal error")
		}
		jsonHelpers.WriteJSON(w, http.StatusOK, UserManagmentResponse{Success: true, Message: "user addedd to the roles"})
	})
}
