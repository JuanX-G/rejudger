package login

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"revit/internal/db"
	"revit/internal/passwords"
	"revit/internal/sharedtesting"
	"revit/internal/testingservices"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

type MockLoginStore struct{}

var PASSWORD, _ = passwords.MakeHash(testingservices.DEFAULT_PASSWORD)

func (ls *MockLoginStore) GetBasicInfoByEmail(context.Context, pgtype.Text) (db.GetBasicInfoByEmailRow, error) {
	out := db.GetBasicInfoByEmailRow{}
	out.Name = testingservices.DEFAULT_USERNAME
	out.ID = testingservices.DEFAULT_USERID
	out.Password = PASSWORD
	return out, nil
}

func (ls *MockLoginStore) GetBasicInfoByUserName(context.Context, string) (db.GetBasicInfoByUserNameRow, error) {
	return db.GetBasicInfoByUserNameRow{
		Name:     testingservices.DEFAULT_USERNAME,
		ID:       testingservices.DEFAULT_USERID,
		Password: PASSWORD,
	}, nil
}

func TestLoginEmail(t *testing.T) {
	_, authmgr, err := testingservices.MakeAuthMgr(testingservices.DEFAULT_PERMISSION_SET)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "auth")
	}

	mux := http.NewServeMux()
	store := &MockLoginStore{}
	permMgr := &testingservices.MockPermissionsManager{}

	service, err := NewLoginManager(t.Context(), store, authmgr, permMgr)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "mock permissions")
	}
	service.RegisterRoutes(mux)

	query := LoginQuery{Password: testingservices.DEFAULT_PASSWORD, Email: testingservices.DEFAULT_EMAIL}
	b, err := json.Marshal(query)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json marshalling of struct LoginQuery")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/session/login", reader)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		sharedtesting.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		sharedtesting.OperationFail(t, err, "io.ReadAll of response body")
	}

	var resp LoginResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json unmarshaling of response body")
	}

	if !resp.Success {
		t.Fatalf("expected the operation to return: %t, found: %t", true, resp.Success)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	_, authmgr, err := testingservices.MakeAuthMgr(testingservices.DEFAULT_PERMISSION_SET)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "auth")
	}

	mux := http.NewServeMux()
	store := &MockLoginStore{}
	permMgr := &testingservices.MockPermissionsManager{}

	service, err := NewLoginManager(t.Context(), store, authmgr, permMgr)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "mock permissions")
	}
	service.RegisterRoutes(mux)

	query := LoginQuery{Password: "word", Email: testingservices.DEFAULT_EMAIL}
	b, err := json.Marshal(query)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json marshalling of struct LoginQuery")
	}

	reader := bytes.NewReader(b)

	req := httptest.NewRequest(http.MethodPost, "/session/login", reader)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		sharedtesting.WrongHttpCode(t, http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		sharedtesting.OperationFail(t, err, "io.ReadAll of response body")
	}

	var resp LoginResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json unmarshaling of response body")
	}

	if resp.Success {
		t.Fatalf("expected the operation to return: %t, found: %t", false, resp.Success)
	}
}
