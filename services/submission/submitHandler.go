package submission

import (
	"context"
	"fmt"
	"net/http"

	"revit/internal/auth"
	"revit/internal/db"
	"revit/internal/hashing"
	"revit/internal/httpHelpers"
)

type SubmissionQuery struct {
	Content    string   `json:"content"`
	AppContext string   `json:"context"`
	Hashes     []string `json:"artifacts"`
}

type SubmissionResponse struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash"`
}

// Submit is now a neutral, transport-agnostic business logic method.
func (ss *SubmissionService) Submit(ctx context.Context, query SubmissionQuery) (SubmissionResponse, error) {
	appCtx, ok := auth.GetAppContextFromCtx(ctx)
	if !ok {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusForbidden, "App contexts do not line up")
	}

	pipeline, err := ss.store.GetPipelineByContextName(ctx, appCtx)
	if err != nil {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "Could not parse request")
	}

	if len(query.Content) == 0 {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusBadRequest, "Content cannot be empty")
	}

	authorID, ok := auth.GetIdFromContext(ctx)
	if !ok {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusUnauthorized, "Unauthorized access")
	}

	user, err := ss.store.GetUserById(ctx, authorID)
	if err != nil {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "User not found")
	}

	hash := hashing.HashSubmission(query.Content, user.Name)

	firstStageId, err := ss.store.GetFirstPipelineStageID(ctx, pipeline.ID)
	if err != nil {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "Internal error")
	}

	err = ss.store.InsertSubmission(ctx, db.InsertSubmissionParams{
		Content:       query.Content,
		Author:        user.ID,
		Hash:          hash,
		Pipeline:      pipeline.ID,
		PipelineStage: firstStageId,
	})

	if err != nil {
		return SubmissionResponse{}, fmt.Errorf("failed to insert submission: %w", err)
	}

	submission, err := ss.store.GetSubmissionsByHash(ctx, hash)
	if err != nil {
		return SubmissionResponse{}, fmt.Errorf("failed to retrieve submission: %w", err)
	}

	ss.artifacts.ExpectArtifacts(query.Hashes, hash, submission.ID)

	return SubmissionResponse{
		Success: true,
		Hash:    hash,
	}, nil
}
