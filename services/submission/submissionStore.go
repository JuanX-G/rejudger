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
	AddPlusOne(ctx context.Context, arg db.AddPlusOneParams) error
	CountPlusOnes(ctx context.Context, arg db.CountPlusOnesParams) (int64, error)
	GetPipelineStage(ctx context.Context, arg db.GetPipelineStageParams) (db.PipelineStage, error)
	GetStageById(ctx context.Context, id int64) (db.Stage, error)
	AddOneToSubmissionStageById(ctx context.Context, id int64) error
	GetVetosBySubmissionId(ctx context.Context, submissionID int64) ([]db.Veto, error)
	DeletePlusOne(ctx context.Context, arg db.DeletePlusOneParams) error
	GetPipelineByContextName(ctx context.Context, name string) (db.Pipeline, error)
	GetFirstPipelineStageID(ctx context.Context, pipelineID int64) (int64, error)
	GetStagePositionByID(ctx context.Context, id int64) (int32, error)
}
