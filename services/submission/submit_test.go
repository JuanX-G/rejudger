package submission

import (
	"revit/internal/hashing"
	"revit/internal/logger"
	shared "revit/internal/sharedtesting"
	services "revit/internal/testingservices"
	"testing"
	"time"
)

func TestSubmit(t *testing.T) {
	token, svcs, cleanup := setUpServices(t)
	defer cleanup()

	insertCh := make(chan InsertionNotification, 8)
	defer close(insertCh)
	svcs.store = &mockSubmissionStore{insertNotify: insertCh, AcceptAll: true}

	ctx := makeFullCtx(t, token)
	ss := NewSubmissionService(SubmissionServiceParams{
		Store:       svcs.store,
		AuthMgr:     svcs.auth,
		ArtifactSvc: svcs.artifacts,
		BaseLogger:  svcs.logger,
		AppContext:  services.DEFAULT_APPCONTEXT,
	})

	query := SubmissionQuery{Content: TEST_SUBMISSION, AppContext: services.DEFAULT_APPCONTEXT}

	resp, err := ss.Submit(ctx, query)

	if err != nil {
		t.Fatalf("expected err == nil, found err: %+v", err)
	}

	const EXPECTED_SUCCESS = true
	if resp.Success != EXPECTED_SUCCESS {
		t.Fatalf("expected success == %t, found success: %t", EXPECTED_SUCCESS, resp.Success)
	}

	inserted := <-insertCh
	if inserted.Content != TEST_SUBMISSION {
		t.Fatalf("expected the content of the inserted submission to be: %s, found: %s", TEST_SUBMISSION, inserted.Content)
	}

	if inserted.Author != services.DEFAULT_USERID {
		t.Fatalf("expected the author id of the inserted submission to be: %d, found: %d", services.DEFAULT_USERID, inserted.Author)
	}
}

func TestSubmitContentEmpty(t *testing.T) {
	token, svcs, cleanup := setUpServices(t)
	defer cleanup()

	insertCh := make(chan InsertionNotification, 8)
	defer close(insertCh)
	svcs.store = &mockSubmissionStore{insertNotify: insertCh, AcceptAll: true}

	ctx := makeFullCtx(t, token)
	ss := NewSubmissionService(SubmissionServiceParams{
		Store:       svcs.store,
		AuthMgr:     svcs.auth,
		ArtifactSvc: svcs.artifacts,
		BaseLogger:  svcs.logger,
		AppContext:  services.DEFAULT_APPCONTEXT,
	})

	query := SubmissionQuery{Content: "", AppContext: services.DEFAULT_APPCONTEXT}

	resp, err := ss.Submit(ctx, query)

	if err == nil {
		t.Fatalf("expected err != nil, found err: %+v", err)
	}

	const EXPECTED_SUCCESS = false
	if resp.Success != EXPECTED_SUCCESS {
		t.Fatalf("expected success == %t, found success: %t", EXPECTED_SUCCESS, resp.Success)
	}

}

func TestSubmitArtifactUploads(t *testing.T) {
	token, svcs, cleanup := setUpServices(t)
	defer cleanup()

	insertCh := make(chan InsertionNotification, 8)
	defer close(insertCh)
	svcs.store = &mockSubmissionStore{insertNotify: insertCh, AcceptAll: true}

	ctx := makeFullCtx(t, token)
	ss := NewSubmissionService(SubmissionServiceParams{
		Store:       svcs.store,
		AuthMgr:     svcs.auth,
		ArtifactSvc: svcs.artifacts,
		BaseLogger:  svcs.logger,
		AppContext:  services.DEFAULT_APPCONTEXT,
	})

	query := SubmissionQuery{Content: TEST_SUBMISSION, Hashes: []string{services.DEFAULT_OBJECT_HASH}}

	resp, err := ss.Submit(ctx, query)
	if err != nil {
		t.Fatalf("expected err == nil, found err: %+v", err)
	}

	if !resp.Success {
		t.Fatalf("expected response to report success, found: %t", resp.Success)
	}

	art := <-svcs.artifacts.ExpectedArtifacts
	if len(art.Hashes) != 1 {
		t.Fatalf("expected 1 artifact hash, found: %d", len(art.Hashes))
	}

	if art.Hashes[0] != services.DEFAULT_OBJECT_HASH {
		t.Fatalf("expected artifact hash: %s, found: %s", services.DEFAULT_OBJECT_HASH, art.Hashes[0])
	}

	hash := hashing.HashSubmission(query.Content, services.DEFAULT_USERNAME)
	if art.SubmissionHash != hash {
		t.Fatalf("expected submission hash tied to the artifact to be: %s, found: %s", hash, art.SubmissionHash)
	}

	inserted := <-insertCh
	if art.SubmissionId != inserted.submissionId {
		t.Fatalf("expected submission id on the artifact to be: %d, found: %d", inserted.submissionId, art.SubmissionId)
	}
}

func TestServiceTimeoutSetting(t *testing.T) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(t.Context(), logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	_, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}

	svc := NewSubmissionService(SubmissionServiceParams{
		AuthMgr:    authmgr,
		BaseLogger: logr,
		AppContext: services.DEFAULT_APPCONTEXT,
		Timeout:    0,
	})
	expected := DEFAULT_SUBMISSION_TIMEOUT * time.Second
	if svc.timeout != expected {
		t.Fatalf("expected timeout to be: %s, found: %s", expected, svc.timeout)
	}
}
