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
	Success             bool `json:"success"`
	SentForPublishing   bool `json:"sent_for_publishing"`
	MovedToTheNextStage bool `json:"moved_to_the_next_stage"`
	HardVetoed          bool `json:"hard_vetoed"`
	SoftVetoed          bool `json:"soft_vetoed"`
}

func (ss *SubmissionService) failPlusOne(ctx context.Context, w http.ResponseWriter, msg string) {
	jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
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
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, AddPlusOneToSubmissionResponse{Success: false})
			return
		}

		ctx, cancel := ss.requestCtx(r)
		defer cancel()

		submission, err := ss.store.GetSubmissionsByHash(ctx, query.Hash)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
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

		currentPlusOnes, err := ss.store.CountPlusOnes(ctx, db.CountPlusOnesParams{SubmissionID: submission.ID, PipelineStage: submission.PipelineStage})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
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

		hardVeto := false
		softVeto := false

		stage, err := ss.store.GetStageById(ctx, stageId.StageID)
		if currentPlusOnes+1 >= int64(stage.PlusOneRequired) {
			vetos, err := ss.store.GetVetosBySubmissionId(ctx, submission.ID)
			if err != nil {
				ss.failPlusOne(ctx, w, "TODO")
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
					jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true, SentForPublishing: true})
					// TODO: send to publishing
					return
				} else if softVeto && !hardVeto {
					pastDeadline, err := ss.deadlineMgr.PastDeadline(ctx, stage.ID)
					if err != nil {
						ss.logger.Log(ctx,
							"TODO!!",
							logger.LogLevelError,
						)
						jsonHelpers.WriteJSON(w, 500, AddPlusOneToSubmissionResponse{Success: false, SentForPublishing: false})
						return
					}
					if pastDeadline {
						err := ss.store.AddOneToSubmissionStageById(ctx, submission.ID)
						if err != nil {
							ss.logger.Log(ctx,
								"TODO!!",
								logger.LogLevelError,
							)
							jsonHelpers.WriteJSON(w, 500, AddPlusOneToSubmissionResponse{Success: false, SentForPublishing: false})
							return
						}
						ss.logger.Log(ctx,
							fmt.Sprintf("submission with hash: %s was moved along to stage number: %d, despite a soft-veto as the deadline has been reached", submission.Hash, submission.PipelineStage+1),
							logger.LogLevelInfo,
						)
						jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true, MovedToTheNextStage: true})
						return
					}

				} else {
					// ERROR!
				}
			} else {
				ss.store.AddOneToSubmissionStageById(ctx, submission.ID)
				ss.logger.Log(ctx,
					fmt.Sprintf("submission with hash: %s was moved along to stage number: %d", submission.Hash, submission.PipelineStage+1),
					logger.LogLevelInfo,
				)
				jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true, MovedToTheNextStage: true})
				return
			}
		}

		err = ss.store.AddPlusOne(ctx, db.AddPlusOneParams{UserID: userid, SubmissionID: submission.ID, PipelineStage: submission.PipelineStage})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
			ss.logger.LogFrom(ctx,
				"add plus one to submission handler",
				fmt.Sprintf("failed to add plusone, error reported: %+v", err),
				logger.LogLevelError,
			)
			return
		}

		if hardVeto || softVeto {
			ss.logger.Log(ctx,
				fmt.Sprintf("submission with hash: %s is vetoed so it did not move to the next stage. Hard veto status: %t", submission.Hash, hardVeto),
				logger.LogLevelInfo,
			)
			jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true, HardVetoed: hardVeto, SoftVetoed: softVeto})
		} else {
			ss.logger.Log(ctx,
				fmt.Sprintf("submission with hash: %s was plusoned, current plusone count: %d", submission.Hash, currentPlusOnes+1),
				logger.LogLevelInfo,
			)
			jsonHelpers.WriteJSON(w, 200, AddPlusOneToSubmissionResponse{Success: true})
		}
		return

	}
}

//
// vetoers := []sharedmodels.Veto{}
// 			vetoerIds := []struct {
// 				id   int64
// 				hard bool
// 			}{}
//
// slices.Grow(vetoers, len(vetos))
// 				slices.Grow(vetoerIds, len(vetos))
//
//
// vetoerIds = append(vetoerIds, struct {
// 						id   int64
// 						hard bool
// 					}{id: v.Vetoer, hard: v.Hard})
//
// for _, v := range vetoerIds {
// 				id := v.id
// 				users, err := ss.store.GetUserById(ctx, id)
// 				if err != nil {
// 					ss.logger.LogFrom(ctx,
// 						"add plus one to submission handler",
// 						fmt.Sprintf("failed to retrive user with id: %d, error reported: %+v", id, err),
// 						logger.LogLevelError,
// 					)
// 					jsonHelpers.WriteJSON(w, http.StatusInternalServerError, AddPlusOneToSubmissionResponse{Success: false})
// 					return
// 				}
//
// 				userData := sharedmodels.MapUserDataFromDBModel(users)
// 				vetoers = append(vetoers, sharedmodels.Veto{Vetoer: userData, Hard: v.hard})
// 			}
