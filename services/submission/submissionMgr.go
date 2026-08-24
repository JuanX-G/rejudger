package submission

import (
	"net/http"
	"revit/services/artifactservice"
	"revit/services/auth"
	"time"
)

type SubmissionService struct {
	guard     *auth.EndpointGuard
	store     submissionStore
	auth      auth.AuthManager
	artifacts artifactservice.ArtifactService
	timeout   time.Duration
}

const DEFAULT_SUBMISSION_TIMEOUT = 10

func NewSubmissionService(queries submissionStore, authMgr auth.AuthManager, artifactSvc artifactservice.ArtifactService, appContext string, timeout uint) *SubmissionService {
	guard := auth.NewEndpointGuard(authMgr, appContext)
	if timeout != 0 {
		return &SubmissionService{guard: guard, store: queries, timeout: time.Duration(timeout) * time.Second, auth: authMgr, artifacts: artifactSvc}
	} else {
		return &SubmissionService{guard: guard, store: queries, timeout: DEFAULT_SUBMISSION_TIMEOUT * time.Second, auth: authMgr, artifacts: artifactSvc}
	}
}

func (s *SubmissionService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /submissions/submit", s.guard.LockSubmit(s.HandleSubmission()))
}
