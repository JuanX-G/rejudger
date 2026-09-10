package submission

import (
	"fmt"
	"net/http"
	"revit/internal/config/sharedmodels"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/internal/logger"
)

type GetSubmissionByHash struct {
	Hash string `json:"hash"`
}

type GetSubmissionResponse struct {
	Success    bool                    `json:"success"`
	Msg        string                  `json:"message"`
	Submission sharedmodels.Submission `json:"submissions"`
}

func (ss *SubmissionService) HandleGetSubmissionHash() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query GetSubmissionByHash
		_, err := httpHelpers.RequestJsonToStruct(r, &query)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, SubmissionResponse{Success: false})
			return
		}

		ctx, cancel := ss.requestCtx(r)

		if len(query.Hash) < 1 {
			getSubmissionFail(w, "recived an empty hash", 400)
			ss.logger.Log(ctx, "error empty hash was sent", logger.LogLevelError)
			return
		}

		defer cancel()

		submission, err := ss.store.GetSubmissionsByHash(ctx, query.Hash)
		if err != nil {
			getSubmissionFail(w, "internal server error", 500)
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %+v", err), logger.LogLevelError)
			return
		}

		author, err := ss.store.GetUserById(ctx, submission.Author)
		if err != nil {
			getSubmissionFail(w, "internal server error", 500)
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %+v", err), logger.LogLevelError)
			return
		}
		var email string
		if author.Email.Valid {
			email = author.Email.String
		}

		resp := GetSubmissionResponse{
			Success: true,
			Msg:     "",
			Submission: sharedmodels.Submission{
				DateCrated: submission.CreatedAt.Time,
				Author:     sharedmodels.UserData{Email: email, Name: author.Name},
				Content:    submission.Content,
				Hash:       submission.Hash,
			},
		}
		jsonHelpers.WriteJSON(w, 200, resp)
	}
}
