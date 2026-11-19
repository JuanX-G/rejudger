package submission

import (
	"context"
	"net/http"
	"revit/internal/auth"
	"revit/internal/logger"
	"revit/internal/sharedmodels"
	shared "revit/internal/sharedtesting"
	services "revit/internal/testingservices"
	"testing"
)

type testServices struct {
	logger    *logger.MultiLogger
	auth      auth.AuthManager
	store     submissionStore
	artifacts *services.MockArtifactService
	mux       *http.ServeMux
}

func setUpServices(t testing.TB) (string, testServices, func()) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(t.Context(), logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	store := &mockSubmissionStore{}

	mux := http.NewServeMux()
	artifactSvc := &services.MockArtifactService{}

	artifactSvc.ExpectedArtifacts = make(chan services.ExpectedMsg, 8)
	cleanupFn := func() {
		close(artifactSvc.ExpectedArtifacts)
	}
	service := NewSubmissionService(SubmissionServiceParams{
		Store:       store,
		AuthMgr:     authmgr,
		ArtifactSvc: artifactSvc,
		BaseLogger:  logr,
		AppContext:  services.DEFAULT_APPCONTEXT,
	})
	service.RegisterRoutes(mux)
	return token, testServices{
		logger:    logr,
		auth:      authmgr,
		store:     store,
		artifacts: artifactSvc,
		mux:       mux,
	}, cleanupFn
}

func setUpTestSubmissionService(t *testing.T) (context.Context, testServices, *SubmissionService) {
	token, svcs, cleanup := setUpServices(t)
	defer cleanup()
	ctx := makeFullCtx(t, token)
	return ctx, svcs, NewSubmissionService(SubmissionServiceParams{
		Store:       svcs.store,
		AuthMgr:     svcs.auth,
		ArtifactSvc: svcs.artifacts,
		BaseLogger:  svcs.logger,
		AppContext:  services.DEFAULT_APPCONTEXT,
	})
}

func makeFullCtx(t *testing.T, token string) context.Context {
	return auth.ContextWithAppCtx(
		auth.ContextWithId(
			auth.ContextWithToken(t.Context(), token),
			DEFAULT_SUBMISSION_1.Author,
		),
		services.DEFAULT_APPCONTEXT,
	)
}

// Test fetching all submissions for a user.
func TestGetSubmission(t *testing.T) {
	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 0}
	ctx, _, ss := setUpTestSubmissionService(t)

	resp, err := ss.GetSubmissionByAuthor(ctx, query)
	if err != nil {
		t.Fatalf("expected err == nil, found err: %+v", err)
	}

	const EXPECTED_SUBS = 2
	if len(resp.Submissions) != EXPECTED_SUBS {
		t.Fatalf("expected to find %d submissions, found: %d", EXPECTED_SUBS, len(resp.Submissions))
	}

	if resp.Submissions[0] != sharedmodels.MapDBSubmissionToModel(DEFAULT_SUBMISSION_1, DEFAULT_USERDATA) {
		t.Fatalf("expected submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_1, resp.Submissions[0])
	}

	if resp.Submissions[1] != sharedmodels.MapDBSubmissionToModel(DEFAULT_SUBMISSION_2, DEFAULT_USERDATA) {
		t.Fatalf("expected submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_2, resp.Submissions[1])
	}
}

func TestGetSubmissionInvalidEmail(t *testing.T) {
	query := GetSubmissionByAuthorQuery{Email: "NOT-AN_EMAIL_HERE11192.FOR_SUR-E99()@NOT_HERE.YOU.WISH", Offset: 0}
	ctx, _, ss := setUpTestSubmissionService(t)

	resp, err := ss.GetSubmissionByAuthor(ctx, query)
	if err == nil {
		t.Fatalf("expected err != nil, found err: %+v", err)
	}

	const EXPECTED_SUBS = 0
	if len(resp.Submissions) != EXPECTED_SUBS {
		t.Fatalf("expected to find %d submissions, found: %d", EXPECTED_SUBS, len(resp.Submissions))
	}

	const EXPECTED_RESULT = false
	if resp.Success != EXPECTED_RESULT {
		t.Fatalf("expected success: %t, found: %t", EXPECTED_RESULT, resp.Success)
	}
}

// Test getting submissions with an offset into the list
func TestGetSubmissionOffset(t *testing.T) {
	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 1}
	ctx, _, ss := setUpTestSubmissionService(t)

	resp, err := ss.GetSubmissionByAuthor(ctx, query)
	if err != nil {
		t.Fatalf("expected err == nil, found err: %+v", err)
	}

	const EXPECTED_SUBS = 1
	if len(resp.Submissions) != EXPECTED_SUBS {
		t.Fatalf("expected to find %d submissions, found: %d", EXPECTED_SUBS, len(resp.Submissions))
	}

	if resp.Submissions[0] != sharedmodels.MapDBSubmissionToModel(DEFAULT_SUBMISSION_2, DEFAULT_USERDATA) {
		t.Fatalf("expected submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_1, resp.Submissions[0])
	}

}

func TestGetSubmissionByHash(t *testing.T) {
	query := GetSubmissionByHash{Hash: DEFAULT_SUBMISSION_1.Hash}
	ctx, _, ss := setUpTestSubmissionService(t)

	resp, err := ss.GetSubmissionByHash(ctx, query)
	if err != nil {
		t.Fatalf("expected err == nil, found err: %+v", err)
	}
	if resp.Submission != DEFAULT_RESP_SUBMISSION_1 {
		t.Fatalf("expected submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_1, resp.Submission)
	}
}

func TestGetSubmissionByEmptyHash(t *testing.T) {
	query := GetSubmissionByHash{Hash: ""}
	ctx, _, ss := setUpTestSubmissionService(t)

	resp, err := ss.GetSubmissionByHash(ctx, query)
	if err == nil {
		t.Fatalf("expected err != nil, found err: %+v", err)
	}

	const EXPECTED_RESULT = false
	if resp.Success != EXPECTED_RESULT {
		t.Fatalf("expected success: %t, found: %t", EXPECTED_RESULT, resp.Success)
	}

	submission := sharedmodels.Submission{}
	if resp.Submission != submission {
		t.Fatalf("expected submission: %+v, found: %+v", submission, resp.Submission)
	}
}
