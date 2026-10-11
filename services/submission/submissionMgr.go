package submission

import (
	"context"
	"net/http"
	"revit/internal/deadlines"
	"revit/internal/jsonHelpers"
	"revit/internal/logger"
	"revit/internal/permissions"
	"revit/services/artifactservice"
	"revit/services/auth"
	"revit/services/publishingservice"
	"time"
)

type SubmissionService struct {
	guard       *auth.EndpointGuard
	store       submissionStore
	auth        auth.AuthManager
	artifacts   artifactservice.ArtifactService
	timeout     time.Duration
	logger      logger.ScopedLogger
	publisher   publishingservice.PublishingService
	deadlineMgr deadlines.DeadlineMgr
}

const DEFAULT_SUBMISSION_TIMEOUT = 10

type SubmissionServiceParams struct {
	Store       submissionStore
	AuthMgr     auth.AuthManager
	ArtifactSvc artifactservice.ArtifactService
	BaseLogger  *logger.MultiLogger
	AppContext  string
	Timeout     uint
	Publisher   publishingservice.PublishingService
}

func NewSubmissionService(params SubmissionServiceParams) *SubmissionService {
	guard := auth.NewEndpointGuard(params.AuthMgr, params.AppContext)
	svc := &SubmissionService{
		guard:     guard,
		store:     params.Store,
		auth:      params.AuthMgr,
		artifacts: params.ArtifactSvc,
		logger:    *logger.NewScopedLogger(params.BaseLogger, "submission-service"),
	}
	if params.Timeout != 0 {
		svc.timeout = time.Duration(params.Timeout) * time.Second
	} else {
		svc.timeout = DEFAULT_SUBMISSION_TIMEOUT * time.Second
	}
	return svc
}

func (ss *SubmissionService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /submissions/submit", ss.guard.Lock(ss.HandleSubmission(), permissions.NewPermissionSet("", permissions.PermissionSubmit)))
	mux.HandleFunc("POST /submissions/get_by_author", ss.guard.Lock(ss.HandleGetSubmissionByAuthor(), permissions.NewPermissionSet("", permissions.PermissionView)))
	mux.HandleFunc("POST /submissions/get_by_author/after", ss.guard.Lock(ss.HandleGetSubmissionAfterByAuthor(), permissions.NewPermissionSet("", permissions.PermissionView)))
	mux.HandleFunc("POST /submissions/get_by_hash", ss.guard.Lock(ss.HandleGetSubmissionByHash(), permissions.NewPermissionSet("", permissions.PermissionView)))
}

func (ss *SubmissionService) requestCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), ss.timeout)
}

func getSubmissionFail(w http.ResponseWriter, msg string, code int) {
	res := GetSubmissionResponse{
		Success: false,
		Msg:     msg,
	}
	jsonHelpers.WriteJSON(w, code, res)
}
