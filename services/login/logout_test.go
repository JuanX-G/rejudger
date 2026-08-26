package login

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"revit/internal/sharedtesting"
	"revit/internal/testingservices"
	"testing"
)

func TestLogout(t *testing.T) {
	token, authMgr, err := testingservices.MakeAuthMgr(testingservices.DEFAULT_PERMISSION_SET)
	if err != nil {

	}

	permMgr := &testingservices.MockPermissionsManager{}

	store := &MockLoginStore{}
	loginMgr, err := NewLoginManager(t.Context(), store, authMgr, permMgr)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "login manager")
	}
	body, code, closeFn := MakeLogoutRequest(t, loginMgr, token)
	defer closeFn()

	var resp logoutResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json unmarshaling of response body")
	}

	if code != http.StatusOK {
		t.Fatalf("expectec code: %d, found: %d", http.StatusOK, code)
	}

	if !resp.Success {
		t.Fatalf("expected logout to succeed, found success: %t", resp.Success)
	}

	if authMgr.DeletedToken != token {
		t.Fatalf("called logout with token: %s, token deleted was: %s", token, authMgr.DeletedToken)
	}
}

func TestLogoutInvalidSession(t *testing.T) {
	_, authMgr, err := testingservices.MakeAuthMgr(testingservices.DEFAULT_PERMISSION_SET)
	if err != nil {

	}

	permMgr := &testingservices.MockPermissionsManager{}

	store := &MockLoginStore{}
	loginMgr, err := NewLoginManager(t.Context(), store, authMgr, permMgr)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "login manager")
	}
	body, code, closeFn := MakeLogoutRequest(t, loginMgr, "NOT_A_SESSION_NEVER_WAS_A_SESSION_HOPEFULLY_PLEASE_DEAR_GAWD_DAWG!")
	defer closeFn()

	var resp logoutResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json unmarshaling of response body")
	}

	if code != http.StatusBadRequest {
		t.Fatalf("expectec code: %d, found: %d", http.StatusBadRequest, code)
	}

	if resp.Success {
		t.Fatalf("expected logout to fail, found success: %t", resp.Success)
	}
}

func TestLogoutWithoutToken(t *testing.T) {
	_, authMgr, err := testingservices.MakeAuthMgr(testingservices.DEFAULT_PERMISSION_SET)
	if err != nil {

	}

	permMgr := &testingservices.MockPermissionsManager{}
	store := &MockLoginStore{}
	loginMgr, err := NewLoginManager(t.Context(), store, authMgr, permMgr)
	if err != nil {
		sharedtesting.ServiceSetupFail(t, err, "login manager")
	}

	mux := http.NewServeMux()
	loginMgr.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/session/logout", nil)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	var resp logoutResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		sharedtesting.OperationFail(t, err, "json unmarshaling of response body")
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expectec code: %d, found: %d", http.StatusOK, res.StatusCode)
	}

	if resp.Success {
		t.Fatalf("expected logout to fail, found success: %t", resp.Success)
	}
}

func MakeLogoutRequest(t *testing.T, loginMgr *LoginManager, token string) ([]byte, int, func()) {
	mux := http.NewServeMux()
	loginMgr.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/session/logout", nil)
	req.Header.Add("X-Auth-Token", token)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		sharedtesting.OperationFail(t, err, "io.ReadAll of response body")
	}
	return bytes, res.StatusCode, func() { res.Body.Close() }
}
