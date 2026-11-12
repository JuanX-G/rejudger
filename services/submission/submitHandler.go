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
	authorID, ok := auth.GetIdFromContext(ctx)
	if !ok {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusUnauthorized, "unauthorized access")
	}

	user, err := ss.store.GetUserById(ctx, authorID)
	if err != nil {
		return SubmissionResponse{}, httpHelpers.NewError(http.StatusInternalServerError, "user not found")
	}

	hash := hashing.HashSubmission(query.Content, user.Name)

	err = ss.store.InsertSubmission(ctx, db.InsertSubmissionParams{
		Content:       query.Content,
		Author:        user.ID,
		Hash:          hash,
		Pipeline:      19,
		PipelineStage: 37,
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
