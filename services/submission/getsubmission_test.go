package submission

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"reflect"
	"revit/internal/config/sharedmodels"
	"revit/internal/logger"
	shared "revit/internal/sharedtesting"
	services "revit/internal/testingservices"
	"revit/services/auth"
	"testing"
)

type testServices struct {
	logger    *logger.MultiLogger
	auth      auth.AuthManager
	store     submissionStore
	artifacts *services.MockArtifactStore
	mux       *http.ServeMux
}

func setUpServices(t testing.TB) (string, testServices) {
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
	return token, testServices{
		logger:    logr,
		auth:      authmgr,
		store:     store,
		artifacts: artifactStore,
		mux:       mux,
	}
}

// Test fetching all submissions for a user.
func TestGetSubmission(t *testing.T) {
	token, svcs := setUpServices(t)

	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 0}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct GetSubmissionByAuthorQuery")
	}

	reader := bytes.NewReader(b)
	req := httptest.NewRequest(http.MethodPost, "/submissions/get", reader)
	req.Header.Add("X-Auth-Token", token)

	checkFn := func(t testing.TB, resp GetSubmissionsResponse) {
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

	shared.HttpTest(t, shared.HttpTestParams[GetSubmissionsResponse]{
		Mux:          svcs.mux,
		Req:          req,
		ExpectedCode: http.StatusOK,
		RespChecker:  checkFn,
	})
}

// Test getting submissions with an offset into the list
func TestGetSubmissionOffset(t *testing.T) {
	token, svcs := setUpServices(t)

	query := GetSubmissionByAuthorQuery{Email: services.DEFAULT_EMAIL, Offset: 1, Size: 1}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct GetSubmissionByAuthorQuery")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/submissions/get", reader)
	req.Header.Add("X-Auth-Token", token)

	checkFn := func(t testing.TB, resp GetSubmissionsResponse) {
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

	shared.HttpTest(t, shared.HttpTestParams[GetSubmissionsResponse]{
		Mux:          svcs.mux,
		Req:          req,
		ExpectedCode: http.StatusOK,
		RespChecker:  checkFn,
	})
}

func testGetSubmissionByHash[R any](t *testing.T, hash, url string, code int, checkFn func(t testing.TB, resp R)) *testServices {
	token, svcs := setUpServices(t)

	query := GetSubmissionByHash{Hash: hash}
	b, err := json.Marshal(query)
	if err != nil {
		shared.OperationFail(t, err, "json marshalling of struct")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, url, reader)
	req.Header.Add("X-Auth-Token", token)

	shared.HttpTest(t, shared.HttpTestParams[R]{
		Mux:          svcs.mux,
		Req:          req,
		ExpectedCode: code,
		RespChecker:  checkFn,
	})
	return &svcs
}

func TestGetSubmissionByHash(t *testing.T) {
	checkFn := func(t testing.TB, resp GetSubmissionResponse) {
		if !resp.Success {
			t.Fatalf("expected response to report success = %t, found: %t", true, resp.Success)
		}

		if resp.Submission != DEFAULT_RESP_SUBMISSION_1 {
			t.Fatalf("expeted submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_1, resp.Submission)
		}
	}
	testGetSubmissionByHash(t, DEFAULT_RESP_SUBMISSION_1.Hash, "/submissions/get_by_hash", 200, checkFn)
}

func TestGetSubmissionByEmptyHash(t *testing.T) {
	checkFn := func(t testing.TB, resp GetSubmissionResponse) {
		if resp.Success {
			t.Fatalf("expected response to report success = %t, found: %t", false, resp.Success)
		}

		empty := sharedmodels.Submission{}
		if resp.Submission != empty {
			t.Fatalf("expeted submission: %+v, found: %+v", DEFAULT_RESP_SUBMISSION_1, resp.Submission)
		}
	}
	testGetSubmissionByHash(t, "", "/submissions/get_by_hash", 400, checkFn)
}
