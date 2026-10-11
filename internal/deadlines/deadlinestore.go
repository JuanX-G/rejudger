package deadlines

import (
	"context"
	"revit/internal/db"
)

// using <T = .list::List<.E>, E = .Comparable> def f(l: T, b: E) = {
// 	operator walk = |idx: usize that [$ in rangeof l], p: T|{
// 		match list::At(l, idx) {
// 			eq E ->
// 		}
// 	}
// }

type DeadlineStore interface {
	InsertStageDeadline(ctx context.Context, arg db.InsertStageDeadlineParams) error
	GetDeadlineByStageId(ctx context.Context, stageID int64) (db.StageDeadline, error)
	SetDeadlineLastRanToNowByStageId(ctx context.Context, stageID int64) error
	GetPipelineByStageId(ctx context.Context, stageID int64) (db.Pipeline, error)
	GetStageById(ctx context.Context, id int64) (db.Stage, error)
	GetSubmissionsByStageId(ctx context.Context, stageID int64) ([]db.Submission, error)
	CountPlusOnes(ctx context.Context, arg db.CountPlusOnesParams) (int64, error)
	GetVetosBySubmissionId(ctx context.Context, submissionID int64) ([]db.Veto, error)
	AdvanceSubmissionToNextStage(ctx context.Context, submissionID int64) (int64, error)
}
