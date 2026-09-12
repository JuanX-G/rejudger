package submission

import (
	"fmt"
	"net/http"
	"revit/internal/config/sharedmodels"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/internal/logger"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type GetSubmissionByAuthorQuery struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	Offset   int    `json:"offset"`
	Size     int    `json:"size"`
}

type GetSubmissionsResponse struct {
	Success     bool                      `json:"success"`
	Msg         string                    `json:"message"`
	Submissions []sharedmodels.Submission `json:"submissions"`
}

func (ss *SubmissionService) HandleGetSubmissionByAuthor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query GetSubmissionByAuthorQuery
		_, err := httpHelpers.RequestJsonToStruct(r, &query)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, SubmissionResponse{Success: false})
			return
		}

		ctx, cancel := ss.requestCtx(r)
		defer cancel()
		var user db.User
		if query.Email != "" {
			user, err = ss.store.GetUserByEmail(ctx, pgtype.Text{String: query.Email})
			if err != nil {
				getSubmissionFail(w, "error failed to find user", 500)
				return
			}
		} else if query.UserName != "" {
			user, err = ss.store.GetUserByName(ctx, query.UserName)
			if err != nil {
				getSubmissionFail(w, "error failed to find user", 500)
				ss.logger.Log(ctx, fmt.Sprintf("error retriving user from the db, error: %+v", err), logger.LogLevelError)
				return
			}
		} else {
			getSubmissionFail(w, "error did not find email or username", 400)
			return
		}

		if query.Offset < 0 {
			query.Offset = 0
		}

		size := httpHelpers.CapInt(query.Size, 100)
		submissions, err := ss.store.GetSubmissionsByAuthorOffset(ctx, db.GetSubmissionsByAuthorOffsetParams{
			Author: user.ID,
			Offset: int32(query.Offset),
			Limit:  int32(size),
		})
		if err != nil {
			getSubmissionFail(w, "internal server error", 500)
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %+v", err), logger.LogLevelError)
			return
		}

		resSubmissions := make([]sharedmodels.Submission, 0, len(submissions))
		for _, s := range submissions {
			userdt := sharedmodels.UserData{}
			if user.Email.Valid {
				userdt.Email = user.Email.String
			}
			userdt.Name = user.Name
			resSubmissions = append(resSubmissions, sharedmodels.Submission{
				DateCrated: s.CreatedAt.Time,
				Author:     userdt,
				Content:    s.Content,
				Hash:       s.Hash,
			})
		}
		resp := GetSubmissionsResponse{
			Success:     true,
			Msg:         "",
			Submissions: resSubmissions,
		}
		jsonHelpers.WriteJSON(w, 200, resp)
	}
}

type GetSubmissionAfterByAuthorQuery struct {
	Email    string    `json:"email"`
	UserName string    `json:"user_name"`
	Offset   int       `json:"offset"`
	Size     int       `json:"size"`
	After    time.Time `json:"time"`
}

func (ss *SubmissionService) HandleGetSubmissionAfterByAuthor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query GetSubmissionAfterByAuthorQuery
		_, err := httpHelpers.RequestJsonToStruct(r, &query)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, SubmissionResponse{Success: false})
			return
		}

		ctx, cancel := ss.requestCtx(r)
		defer cancel()
		var user db.User
		if query.Email != "" {
			user, err = ss.store.GetUserByEmail(ctx, pgtype.Text{String: query.Email})
			if err != nil {
				getSubmissionFail(w, "error failed to find user", 500)
				return
			}
		} else if query.UserName != "" {
			user, err = ss.store.GetUserByName(ctx, query.UserName)
			if err != nil {
				getSubmissionFail(w, "error failed to find user", 500)
				ss.logger.Log(ctx, fmt.Sprintf("error retriving user from the db, error: %+v", err), logger.LogLevelError)
				return
			}
		} else {
			getSubmissionFail(w, "error did not find email or username", 400)
			return
		}

		if query.Offset < 0 {
			query.Offset = 0
		}

		size := httpHelpers.CapInt(query.Size, 100)
		submissions, err := ss.store.GetSubmissionsByAuthorOffset(ctx, db.GetSubmissionsByAuthorOffsetParams{
			Author: user.ID,
			Offset: int32(query.Offset),
			Limit:  int32(size),
		})
		if err != nil {
			getSubmissionFail(w, "internal server error", 500)
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %+v", err), logger.LogLevelError)
			return
		}

		resSubmissions := make([]sharedmodels.Submission, 0, len(submissions))
		for _, s := range submissions {
			if !s.CreatedAt.Time.After(query.After) {
				continue
			}

			userdt := sharedmodels.UserData{}
			if user.Email.Valid {
				userdt.Email = user.Email.String
			}

			userdt.Name = user.Name
			resSubmissions = append(resSubmissions, sharedmodels.Submission{
				DateCrated: s.CreatedAt.Time,
				Author:     userdt,
				Content:    s.Content,
				Hash:       s.Hash,
			})
		}
		resp := GetSubmissionsResponse{
			Success:     true,
			Msg:         "",
			Submissions: resSubmissions,
		}
		jsonHelpers.WriteJSON(w, 200, resp)
	}
}
