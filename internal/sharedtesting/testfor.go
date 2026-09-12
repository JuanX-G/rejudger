package sharedtesting

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HttpTestParams[R any] struct {
	Mux          *http.ServeMux
	Req          *http.Request
	ExpectedCode int
	RespChecker  func(testing.TB, R)
}

func HttpTest[R any](t *testing.T, params HttpTestParams[R]) {
	rec := httptest.NewRecorder()

	params.Mux.ServeHTTP(rec, params.Req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != params.ExpectedCode {
		t.Fatalf("expected code: %d, found: %d", params.ExpectedCode, res.StatusCode)
	}

	var query R
	body, err := io.ReadAll(res.Body)
	if err != nil {
		OperationFail(t, err, "ReadAll on response body")
	}

	err = json.Unmarshal(body, &query)
	if err != nil {
		OperationFail(t, err, "Json unmarshall on response body")
	}

	params.RespChecker(t, query)
}
