package submission

import (
	"context"
	"fmt"
	"net/http"
	"revit/internal/auth"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/logger"
)

type RemovePlusOneFromeSubmission struct {
	Hash string `json:"hash"`
}

type RemovePlusOneFromSubmissionResponse struct {
	Success bool `json:"success"`
}

// AddPlusOneToSubmission handles business logic for adding a plus-one to a submission.
func (ss *SubmissionService) RemovePlusOneFromSubmission(ctx context.Context, query AddPlusOneToSubmission) (RemovePlusOneFromSubmissionResponse, error) {
	userID, ok := auth.GetIdFromContext(ctx)
	if !ok {
		return RemovePlusOneFromSubmissionResponse{}, httpHelpers.NewError(http.StatusUnauthorized, "unauthorized access")
	}

	submission, err := ss.store.GetSubmissionsByHash(ctx, query.Hash)
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to retrive submission with hash: %s, error reported: %+v", query.Hash, err),
			logger.LogLevelError,
		)
		return RemovePlusOneFromSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve submission")
	}

	position, err := ss.store.GetStagePositionByID(ctx, submission.PipelineStage)
	if err != nil {
		// TODO: log
		return RemovePlusOneFromSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve pipeline stage")
	}

	stageId, err := ss.store.GetPipelineStage(ctx, db.GetPipelineStageParams{
		PipelineID: submission.Pipeline,
		Position:   position,
	})
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to retrive stage for pipeline with id: %d at position: %d, error reported: %+v", submission.Pipeline, submission.PipelineStage, err),
			logger.LogLevelError,
		)
		return RemovePlusOneFromSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve pipeline stage")
	}

	ss.store.DeletePlusOne(ctx, db.DeletePlusOneParams{UserID: userID, SubmissionID: submission.ID, PipelineStage: stageId.ID})
	return RemovePlusOneFromSubmissionResponse{Success: true}, nil
}
