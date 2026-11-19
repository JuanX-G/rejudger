package submission

import (
	"context"
	"fmt"
	"net/http"

	"revit/internal/auth"
	"revit/internal/db"
	"revit/internal/httpHelpers"
	"revit/internal/logger"
	"revit/internal/sharedmodels"
)

type AddPlusOneToSubmission struct {
	Hash string `json:"hash"`
}

type AddPlusOneToSubmissionResponse struct {
	Success             bool `json:"success"`
	SentForPublishing   bool `json:"sent_for_publishing"`
	MovedToTheNextStage bool `json:"moved_to_the_next_stage"`
	HardVetoed          bool `json:"hard_vetoed"`
	SoftVetoed          bool `json:"soft_vetoed"`
}

// AddPlusOneToSubmission handles business logic for adding a plus-one to a submission.
func (ss *SubmissionService) AddPlusOneToSubmission(ctx context.Context, query AddPlusOneToSubmission) (AddPlusOneToSubmissionResponse, error) {
	submission, err := ss.store.GetSubmissionsByHash(ctx, query.Hash)
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to retrive submission with hash: %s, error reported: %+v", query.Hash, err),
			logger.LogLevelError,
		)
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve submission")
	}

	userid, ok := auth.GetIdFromContext(ctx)
	if !ok {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			"user id missing in the request context",
			logger.LogLevelError,
		)
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "user id missing in context")
	}

	position, err := ss.store.GetStagePositionByID(ctx, submission.PipelineStage)
	if err != nil {
		// TODO: log
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve pipeline stage")
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
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve pipeline stage")
	}

	currentPlusOnes, err := ss.store.CountPlusOnes(ctx, db.CountPlusOnesParams{
		SubmissionID:  submission.ID,
		PipelineStage: stageId.ID,
	})
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to retrive plus ones for submission with hash: %s, error reported: %+v", query.Hash, err),
			logger.LogLevelError,
		)
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to count plus ones")
	}

	stage, err := ss.store.GetStageById(ctx, stageId.StageID)
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to retrive stage with id: %d, error reported: %+v", stageId.StageID, err),
			logger.LogLevelError,
		)
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to retrieve stage")
	}

	hardVeto := false
	softVeto := false

	if currentPlusOnes+1 >= int64(stage.PlusOneRequired) {
		vetos, err := ss.store.GetVetosBySubmissionId(ctx, submission.ID)
		if err != nil {
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				fmt.Sprintf("could not get the veto count for submission with id: %d. Error reported: %s", submission.ID, err),
				logger.LogLevelError,
			)
			return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to get vetos")
		}

		if len(vetos) > 0 {
			for _, v := range vetos {
				if !softVeto {
					softVeto = true
				}
				if v.Hard && !hardVeto {
					hardVeto = true
				}
			}
		}

		if stage.Final {
			if !hardVeto && !softVeto {
				ss.logger.Log(ctx,
					fmt.Sprintf("submission with hash: %s sent of for publishing", submission.Hash),
					logger.LogLevelInfo,
				)

				var authorData sharedmodels.UserData
				author, err := ss.store.GetUserById(ctx, submission.Author)
				if err != nil {
					ss.logger.Log(ctx,
						fmt.Sprintf("Could not retrive user with id: %d", submission.Author),
						logger.LogLevelError,
					)
				}

				if author.Email.Valid {
					authorData.Email = author.Email.String
				}
				authorData.Name = author.Name

				submissionData := sharedmodels.MapDBSubmissionToModel(submission, authorData)

				err = ss.publishing.Publish(submissionData)
				if err != nil {
					// TODO: log
					return AddPlusOneToSubmissionResponse{Success: false, SentForPublishing: false}, nil
				} else {
					return AddPlusOneToSubmissionResponse{Success: true, SentForPublishing: true}, nil
				}
			} else if softVeto && !hardVeto {
				pastDeadline, err := ss.deadlineMgr.PastDeadline(ctx, stage.ID)
				if err != nil {
					ss.logger.Log(ctx,
						fmt.Sprintf("Error checking the deadline of submission with hash: %s. Error reported: %s", submission.Hash, err.Error()),
						logger.LogLevelError,
					)
					return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to check deadline")
				}

				if pastDeadline {
					err := ss.store.AddOneToSubmissionStageById(ctx, submission.ID)
					if err != nil {
						ss.logger.Log(ctx,
							fmt.Sprintf("Error adding a plus one to the submission with hash: %s. Error reported: %s", submission.Hash, err.Error()),
							logger.LogLevelError,
						)
						return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to advance submission stage")
					}
					ss.logger.Log(ctx,
						fmt.Sprintf("submission with hash: %s was moved along to stage number: %d, despite a soft-veto as the deadline has been reached", submission.Hash, submission.PipelineStage+1),
						logger.LogLevelInfo,
					)
					return AddPlusOneToSubmissionResponse{Success: true, MovedToTheNextStage: true}, nil
				}
			} else {
				ss.logger.Log(ctx,
					fmt.Sprintf("submission with hash: %s and ID: %d, was not impacted by a plus one becuase it is hard vetoed", submission.Hash, submission.ID),
					logger.LogLevelInfo,
				)
				return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "submission is hard vetoed")
			}
		} else {
			if err := ss.store.AddOneToSubmissionStageById(ctx, submission.ID); err != nil {
				ss.logger.Log(ctx,
					fmt.Sprintf("Error adding a plus one to the submission with hash: %s. Error reported: %s", submission.Hash, err.Error()),
					logger.LogLevelError,
				)
				return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to advance submission stage")
			}
			ss.logger.Log(ctx,
				fmt.Sprintf("submission with hash: %s was moved along to stage number: %d", submission.Hash, submission.PipelineStage+1),
				logger.LogLevelInfo,
			)
			return AddPlusOneToSubmissionResponse{Success: true, MovedToTheNextStage: true}, nil
		}
	}

	//TODO: Check if this is correct (a.k.a. press x to doubt[?])

	err = ss.store.AddPlusOne(ctx, db.AddPlusOneParams{
		UserID:        userid,
		SubmissionID:  submission.ID,
		PipelineStage: stageId.StageID,
	})
	if err != nil {
		ss.logger.LogFrom(ctx,
			"add plus one to submission handler",
			fmt.Sprintf("failed to add plusone, error reported: %+v", err),
			logger.LogLevelError,
		)
		return AddPlusOneToSubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "failed to add plus one")
	}

	if hardVeto || softVeto {
		ss.logger.Log(ctx,
			fmt.Sprintf("submission with hash: %s is vetoed so it did not move to the next stage. Hard veto status: %t", submission.Hash, hardVeto),
			logger.LogLevelInfo,
		)
		return AddPlusOneToSubmissionResponse{Success: true, HardVetoed: hardVeto, SoftVetoed: softVeto}, nil
	}

	ss.logger.Log(ctx,
		fmt.Sprintf("submission with hash: %s was plusoned, current plusone count: %d", submission.Hash, currentPlusOnes+1),
		logger.LogLevelInfo,
	)
	return AddPlusOneToSubmissionResponse{Success: true}, nil
}
