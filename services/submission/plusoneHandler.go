package submission

import (
	"context"
	"fmt"
	"net/http"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/internal/logger"
	"revit/services/auth"
)

type AddPlusOneToSubmission struct {
	Hash string `json:"hash"`
}

type AddPlusOneToSubmissionResponse struct {
	Success bool `json:"success"`
}

func (ss *SubmissionService) failPlusOne(ctx context.Context, w http.ResponseWriter, msg string) {
	jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
	ss.logger.LogFrom(ctx,
		"'add plus one to submission' handler",
		msg,
		logger.LogLevelError,
	)
}

func (ss *SubmissionService) AddPlusOneToSubmissionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query AddPlusOneToSubmission
		_, err := httpHelpers.RequestJsonToStruct(r, &query)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, SubmissionResponse{Success: false})
			return
		}

		ctx, cancel := ss.requestCtx(r)
		defer cancel()

		submission, err := ss.store.GetSubmissionsByHash(ctx, query.Hash)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				fmt.Sprintf("failed to retrive submission with hash: %s, error reported: %+v", query.Hash, err),
				logger.LogLevelError,
			)
			return
		}

		userid, ok := auth.GetIdFromContext(r.Context())
		if !ok {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				"user id missing in the request context",
				logger.LogLevelError,
			)
		}

		currentPlusOnes, err := ss.store.CountPlusOnes(ctx, submission.ID)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				fmt.Sprintf("failed to retrive plus ones for submission with hash: %s, error reported: %+v", query.Hash, err),
				logger.LogLevelError,
			)
			return
		}

		stageId, err := ss.store.GetPipelineStage(ctx, db.GetPipelineStageParams{PipelineID: submission.Pipeline, Position: submission.PipelineStage})
		if err != nil {
			ss.failPlusOne(ctx, w, fmt.Sprintf("failed to retrive stage, for pipeline with id: %d, at position: %d, error reported: %+v", submission.Pipeline, submission.PipelineStage, err))
		}

		stage, err := ss.store.GetStageById(ctx, stageId.StageID)
		if currentPlusOnes+1 >= int64(stage.PlusOneRequired) {
			if stage.Final {
				// Call the publisher and send the submission
				ss.logger.Log(ctx,
					fmt.Sprintf("submission with hash: %s, sent of for publishing", submission.Hash),
					logger.LogLevelInfo,
				)
				jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true})
			} else {
				// Move the submission along
			}
		}

		err = ss.store.AddPlusOne(ctx, db.AddPlusOneParams{UserID: userid, SubmissionID: submission.ID})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				fmt.Sprintf("failed to add plusone, error reported: %+v", err),
				logger.LogLevelError,
			)
			return
		}

	}
}
