package submission

import (
	"encoding/json/v2"
	"net/http"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type GetSubmissionByAuthorQuery struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	Offset   int    `json:"offset"`
}

type UserData struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GetSubmissionResponse struct {
	Success     bool         `json:"success"`
	Msg         string       `json:"message"`
	Submissions []Submission `json:"submissions"`
}

type Submission struct {
	DateCrated time.Time `json:"date_created"`
	Author     UserData  `json:"author"`
	Content    string    `json:"content"`
	Hash       string    `json:"hash"`
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
			user, err = ss.store.GetUserByUserName(ctx, query.UserName)
			if err != nil {
				getSubmissionFail(w, "error failed to find user", 500)
				return
			}
		} else {
			getSubmissionFail(w, "error did not find email or username", 400)
			return
		}

		if query.Offset < 0 {
			query.Offset = 0
		}
		submissions, err := ss.store.GetSubmissionsByAuthorOffset(ctx, db.GetSubmissionsByAuthorOffsetParams{Author: user.ID, Offset: int32(query.Offset)})
		if err != nil {
			getSubmissionFail(w, "internal server error", 500)
			return
		}

		resSubmissions := make([]Submission, 0, len(submissions))
		for _, s := range submissions {
			userdt := UserData{}
			if user.Email.Valid {
				userdt.Email = user.Email.String
			} else {
				userdt.Name = user.Name
			}
			resSubmissions = append(resSubmissions, Submission{
				DateCrated: s.CreatedAt.Time,
				Author:     userdt,
				Content:    s.Content,
				Hash:       s.Hash,
			})
		}
		resp := GetSubmissionResponse{
			Success:     true,
			Msg:         "",
			Submissions: resSubmissions,
		}
		jsonHelpers.WriteJSON(w, 200, resp)
	}
}

func getSubmissionFail(w http.ResponseWriter, msg string, code int) {
	res := GetSubmissionResponse{
		Success: false,
		Msg:     msg,
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "could not marshall json", 500)
	}
	jsonHelpers.WriteJSON(w, code, bytes)
}
