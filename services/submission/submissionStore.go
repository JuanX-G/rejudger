package submission

import (
	"context"
	"revit/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type submissionStore interface {
	GetUserById(context.Context, int64) (db.User, error)
	GetUserByEmail(ctx context.Context, email pgtype.Text) (db.User, error)
	GetUserByName(ctx context.Context, name string) (db.User, error)
	GetSubmissionsByAuthorOffset(ctx context.Context, arg db.GetSubmissionsByAuthorOffsetParams) ([]db.Submission, error)
	InsertSubmission(context.Context, db.InsertSubmissionParams) error
	GetSubmissionsByHash(context.Context, string) (db.Submission, error)
	GetSubmissionsByAuthor(ctx context.Context, author int64) ([]db.Submission, error)
}
