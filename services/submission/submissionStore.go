package submission

import (
	"context"
	"revit/internal/db"
)

type submissionStore interface {
	GetUserById(context.Context, int64) (db.User, error)
	InsertSubmission(context.Context, db.InsertSubmissionParams) error
	GetSubmissionsByHash(context.Context, string) (db.Submission, error)
}
