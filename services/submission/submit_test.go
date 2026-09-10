package submission

import (
	"bytes"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"revit/internal/config/sharedmodels"
	"revit/internal/hashing"
	"revit/internal/logger"
	shared "revit/internal/sharedtesting"
	services "revit/internal/testingservices"
	"testing"
	"time"
)

func TestSubmitHandler(t *testing.T) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	insertCh := make(chan InsertionNotification, 8)
	defer close(insertCh)
	store := &mockSubmissionStore{insertNotify: insertCh}

	mux := http.NewServeMux()
	artifactStore := &services.MockArtifactStore{}

	artifactService := services.NewMockArtifactService(artifactStore)

	service := NewSubmissionService(store, authmgr, artifactService, logr, services.DEFAULT_APPCONTEXT, 0)
	service.RegisterRoutes(mux)

	// stage a request with no artifacts
	query := SubmissionQuery{Content: TEST_SUBMISSION}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct Submission Query")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/submissions/submit", reader)
	req.Header.Add("X-Auth-Token", token)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req) // run the request

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		shared.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		shared.OperationFail(t, err, "io.ReadAll of response body")
	}
	var resp SubmissionResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		shared.OperationFail(t, err, "json unmarshaling of response body")
	}
	hash := hashing.HashSubmission(query.Content, services.DEFAULT_USERNAME)
	if resp.Hash != hash {
		t.Fatalf("expected hash in response to be: %s, found: %s", hash, resp.Hash)
	}
	if !resp.Success {
		t.Fatalf("expected response to report success, found: %t", resp.Success)
	}
	inserted := <-insertCh
	if inserted.Content != TEST_SUBMISSION {
		t.Fatalf("expected the content of the inserted submission to be: %s, found: %s", TEST_SUBMISSION, inserted.Content)
	}
	if inserted.Hash != hash {
		t.Fatalf("expected the hash of the inserted submission to be: %s, found: %s", hash, inserted.Hash)
	}
	if inserted.Author != services.DEFAULT_USERID {
		t.Fatalf("expected the author id of the inserted submission to be: %d, found: %d", services.DEFAULT_USERID, inserted.Author)
	}

}

func TestSubmitArtifactUploads(t *testing.T) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	insertCh := make(chan InsertionNotification, 8)
	defer close(insertCh)
	store := &mockSubmissionStore{insertNotify: insertCh}

	mux := http.NewServeMux()
	artifactStore := &services.MockArtifactStore{}

	artifactService := services.NewMockArtifactService(artifactStore)
	artifactService.ExpectedArtifacts = make(chan services.ExpectedMsg, 8)
	defer close(artifactService.ExpectedArtifacts)
	service := NewSubmissionService(store, authmgr, artifactService, logr, services.DEFAULT_APPCONTEXT, 0)
	service.RegisterRoutes(mux)

	// stage a request with a artifact
	query := SubmissionQuery{Content: TEST_SUBMISSION, Hashes: []string{services.DEFAULT_OBJECT_HASH}}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct Submission Query")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/submissions/submit", reader)
	req.Header.Add("X-Auth-Token", token)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req) // run the request

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		shared.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	var resp SubmissionResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		shared.OperationFail(t, err, "json unmarshaling of response body")
	}

	if !resp.Success {
		t.Fatalf("expected response to report success, found: %t", resp.Success)
	}

	art := <-artifactService.ExpectedArtifacts
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
	logr, err := logger.MakeDefaultMultiLogger(logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	_, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}

	svc := NewSubmissionService(nil, authmgr, nil, logr, "", 0) // It is okay to use nil for some interfaces provided we do not make any requests

	expected := DEFAULT_SUBMISSION_TIMEOUT * time.Second
	if svc.timeout != expected {
		t.Fatalf("expected timeout to be: %s, found: %s", expected, svc.timeout)
	}

	svc = NewSubmissionService(nil, authmgr, nil, logr, "", 15)

	expected = 15 * time.Second
	if svc.timeout != expected {
		t.Fatalf("expected timeout to be: %s, found: %s", expected, svc.timeout)
	}
}

// Test fetching all submissions for a user.
func TestGetSubmission(t *testing.T) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	store := &mockSubmissionStore{}

	mux := http.NewServeMux()
	artifactStore := &services.MockArtifactStore{}

	artifactService := services.NewMockArtifactService(artifactStore)
	artifactService.ExpectedArtifacts = make(chan services.ExpectedMsg, 8)
	defer close(artifactService.ExpectedArtifacts)
	service := NewSubmissionService(store, authmgr, artifactService, logr, services.DEFAULT_APPCONTEXT, 0)
	service.RegisterRoutes(mux)

	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 0}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct GetSubmissionByAuthorQuery")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/submissions/get", reader)
	req.Header.Add("X-Auth-Token", token)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		shared.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		shared.OperationFail(t, err, "io.ReadAll of response body")
	}
	var resp GetSubmissionsResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		shared.OperationFail(t, err, "json unmarshaling of response body")
	}
	if !resp.Success {
		t.Fatalf("expected response to report success = %t, found: %t", true, resp.Success)
	}
	if len(resp.Submissions) != 2 {
		t.Fatalf("sent offset = %d, email = %s, found: %d submission returned", query.Offset, query.Email, len(resp.Submissions))
	}
	if !reflect.DeepEqual(resp.Submissions, DEFAULT_RESP_SUBMISSIONS) {
		t.Fatalf("found submissions: %+v, expected: %+v", resp.Submissions, DEFAULT_RESP_SUBMISSIONS)
	}
}

// Test getting submissions with an offset into the list
func TestGetSubmissionOffset(t *testing.T) {
	logDest := &services.MockLogDest{}
	logr, err := logger.MakeDefaultMultiLogger(logDest)
	if err != nil {
		shared.ServiceSetupFail(t, err, "logger")
	}
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	store := &mockSubmissionStore{}

	mux := http.NewServeMux()
	artifactStore := &services.MockArtifactStore{}

	artifactService := services.NewMockArtifactService(artifactStore)
	artifactService.ExpectedArtifacts = make(chan services.ExpectedMsg, 8)
	defer close(artifactService.ExpectedArtifacts)
	service := NewSubmissionService(store, authmgr, artifactService, logr, services.DEFAULT_APPCONTEXT, 0)
	service.RegisterRoutes(mux)

	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 1}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct GetSubmissionByAuthorQuery")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/submissions/get", reader)
	req.Header.Add("X-Auth-Token", token)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		shared.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		shared.OperationFail(t, err, "io.ReadAll of response body")
	}
	var resp GetSubmissionsResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		shared.OperationFail(t, err, "json unmarshaling of response body")
	}
	if !resp.Success {
		t.Fatalf("expected response to report success = %t, found: %t", true, resp.Success)
	}
	if len(resp.Submissions) != 1 {
		t.Fatalf("sent offset = %d, email = %s, found: %d submission returned", query.Offset, query.Email, len(resp.Submissions))
	}
	if !reflect.DeepEqual(resp.Submissions, []sharedmodels.Submission{DEFAULT_RESP_SUBMISSION_2}) {
		t.Fatalf("found submissions: %+v, expected: %+v", resp.Submissions, DEFAULT_RESP_SUBMISSIONS)
	}
}
