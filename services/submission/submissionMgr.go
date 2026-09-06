package submission

import (
	"context"
	"net/http"
	"revit/internal/logger"
	"revit/internal/permissions"
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
	logger    logger.ScopedLogger
}

const DEFAULT_SUBMISSION_TIMEOUT = 10

func NewSubmissionService(queries submissionStore, authMgr auth.AuthManager, artifactSvc artifactservice.ArtifactService, baseLogger *logger.MultiLogger, appContext string, timeout uint) *SubmissionService {
	guard := auth.NewEndpointGuard(authMgr, appContext)
	svc := &SubmissionService{
		guard:     guard,
		store:     queries,
		auth:      authMgr,
		artifacts: artifactSvc,
		logger:    *logger.NewScopedLogger(baseLogger, "submission-service"),
	}
	if timeout != 0 {
		svc.timeout = time.Duration(timeout) * time.Second
	} else {
		svc.timeout = DEFAULT_SUBMISSION_TIMEOUT * time.Second
	}
	return svc
}

func (ss *SubmissionService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /submissions/submit", ss.guard.Lock(ss.HandleSubmission(), permissions.NewPermissionSet("", permissions.PermissionSubmit)))
	mux.HandleFunc("POST /submissions/get", ss.guard.Lock(ss.HandleGetSubmissionByAuthor(), permissions.NewPermissionSet("", permissions.PermissionView)))
}

func (ss *SubmissionService) requestCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), ss.timeout)
}
