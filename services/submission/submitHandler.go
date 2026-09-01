package submission

import (
	"net/http"

	"revit/internal/db"
	"revit/internal/hashing"
	"revit/internal/httpHelpers"
	"revit/internal/jsonHelpers"
	"revit/services/auth"
)

type SubmissionQuery struct {
	Content string   `json:"content"`
	Hashes  []string `json:"artifacts"`
}

type SubmissionResponse struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash"`
}

func (s *SubmissionService) HandleSubmission() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.TokenFromContext(r.Context())
		if !ok {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			return
		}

		ctx, cancel := s.requestCtx(r)
		defer cancel()

		id, err := s.auth.GetUserId(token)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			return
		}
		user, err := s.store.GetUserById(ctx, id)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			return
		}

		var query SubmissionQuery
		_, err = httpHelpers.RequestJsonToStruct(r, &query)
		defer r.Body.Close()
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusBadRequest, SubmissionResponse{Success: false})
			return
		}

		hash := hashing.HashSubmission(query.Content, user.Name)

		err = s.store.InsertSubmission(ctx, db.InsertSubmissionParams{Content: query.Content, Author: user.ID, Hash: hash})
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			return
		}
		submission, err := s.store.GetSubmissionsByHash(ctx, hash)
		if err != nil {
			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, SubmissionResponse{Success: false})
			return
		}

		s.artifacts.ExpectArtifacts(query.Hashes, hash, submission.ID)
		jsonHelpers.WriteJSON(w, http.StatusOK, SubmissionResponse{Success: true, Hash: hash})
	}
}
