package submission

import (
	"context"
	"fmt"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/logger"
	"revit/internal/sharedmodels"
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

func (ss *SubmissionService) GetSubmissionByAuthor(ctx context.Context, req GetSubmissionByAuthorQuery) (GetSubmissionsResponse, error) {
	var user db.User
	var err error
	if req.Email != "" {
		user, err = ss.store.GetUserByEmail(ctx, pgtype.Text{String: req.Email})
		if err != nil {
			ss.logger.Log(ctx, fmt.Sprintf("error retriving user from the db, error: %s", err.Error()), logger.LogLevelError)
			return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
		}
	} else if req.UserName != "" {
		user, err = ss.store.GetUserByName(ctx, req.UserName)
		if err != nil {
			ss.logger.Log(ctx, fmt.Sprintf("error retriving user from the db, error: %s", err.Error()), logger.LogLevelError)
			return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
		}
	} else {
		fmt.Printf("req: %+v", req)
		return GetSubmissionsResponse{}, httpHelpers.NewError(400, "invalid hash given")
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	size := httpHelpers.CapInt(req.Size, 100)
	submissions, err := ss.store.GetSubmissionsByAuthorOffset(ctx, db.GetSubmissionsByAuthorOffsetParams{
		Author: user.ID,
		Offset: int32(req.Offset),
		Limit:  int32(size),
	})
	if err != nil {
		ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
		return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
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
	return resp, nil
}

type GetSubmissionAfterByAuthorQuery struct {
	Email    string    `json:"email"`
	UserName string    `json:"user_name"`
	Offset   int       `json:"offset"`
	Size     int       `json:"size"`
	After    time.Time `json:"time"`
}

func (ss *SubmissionService) GetSubmissionAfterByAuthor(ctx context.Context, req GetSubmissionAfterByAuthorQuery) (GetSubmissionsResponse, error) {
	var user db.User
	var err error
	if req.Email != "" {
		user, err = ss.store.GetUserByEmail(ctx, pgtype.Text{String: req.Email})
		if err != nil {
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
			return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
		}
	} else if req.UserName != "" {
		user, err = ss.store.GetUserByName(ctx, req.UserName)
		if err != nil {
			ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
			return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
		}
	} else {
		ss.logger.Log(ctx, "error retriving submissions from the db, no user name or email sent", logger.LogLevelError)
		return GetSubmissionsResponse{}, httpHelpers.NewError(500, "")
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	size := httpHelpers.CapInt(req.Size, 100)
	submissions, err := ss.store.GetSubmissionsByAuthorOffset(ctx, db.GetSubmissionsByAuthorOffsetParams{
		Author: user.ID,
		Offset: int32(req.Offset),
		Limit:  int32(size),
	})
	if err != nil {
		ss.logger.Log(ctx, fmt.Sprintf("error retriving submissions from the db, error: %s", err.Error()), logger.LogLevelError)
		return GetSubmissionsResponse{}, httpHelpers.NewError(500, "could not retrive the submission")
	}

	resSubmissions := make([]sharedmodels.Submission, 0, len(submissions))
	for _, s := range submissions {
		if !s.CreatedAt.Time.After(req.After) {
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
	return resp, nil
}
