package submission

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"revit/internal/db"
	"revit/internal/hashing"
	shared "revit/internal/sharedtesting"
	services "revit/internal/testingservices"
	"testing"
)

type mockSubmissionStore struct {
	insertNotify chan db.InsertSubmissionParams
}

func (ss *mockSubmissionStore) GetUserById(ctx context.Context, id int64) (db.User, error) {
	return db.User{
		ID:   services.DEFAULT_USERID,
		Name: services.DEFAULT_USERNAME,
	}, nil
}

func (ss *mockSubmissionStore) InsertSubmission(ctx context.Context, arg db.InsertSubmissionParams) error {
	ss.insertNotify <- arg
	return nil
}

func (ss *mockSubmissionStore) GetSubmissionsByHash(ctx context.Context, hash string) (db.Submission, error) {
	return db.Submission{
		ID: 123,
	}, nil
}

const TEST_SUBMISSION = "Lorem Ipsum"

func TestSubmitHandler(t *testing.T) {
	token, authmgr, err := services.MakeAuthMgr(services.DEFAULT_PERMISSION_SET)
	if err != nil {
		shared.ServiceSetupFail(t, err, "auth")
	}
	insertCh := make(chan db.InsertSubmissionParams, 8)
	defer close(insertCh)
	store := &mockSubmissionStore{insertNotify: insertCh}

	mux := http.NewServeMux()

	artifactStore := &services.MockArtifactService{}

	service := NewSubmissionService(store, authmgr, artifactStore, services.DEFAULT_APPCONTEXT, 0)
	service.RegisterRoutes(mux)

	// stage a request with not artifacts
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
