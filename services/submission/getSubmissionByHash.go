package submission

import (
	"context"
	"fmt"
	"revit/internal/httpHelpers"
	"revit/internal/logger"
	"revit/internal/sharedmodels"
)

type GetSubmissionByHash struct {
	Hash string `json:"hash"`
}

type GetSubmissionResponse struct {
	Success    bool                    `json:"success"`
	Msg        string                  `json:"message"`
	Submission sharedmodels.Submission `json:"submissions"`
}

func (ss *SubmissionService) GetSubmissionByHash(ctx context.Context, req GetSubmissionByHash) (GetSubmissionResponse, error) {
	if len(req.Hash) < 1 {
		ss.logger.Log(ctx, "error empty hash was sent", logger.LogLevelError)
		return GetSubmissionResponse{}, httpHelpers.NewError(400, "invalid hash given")
	}

	submission, err := ss.store.GetSubmissionsByHash(ctx, req.Hash)
	if err != nil {
		ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
		return GetSubmissionResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
	}

	author, err := ss.store.GetUserById(ctx, submission.Author)
	if err != nil {
		ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
		return GetSubmissionResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
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
	return resp, nil
}
