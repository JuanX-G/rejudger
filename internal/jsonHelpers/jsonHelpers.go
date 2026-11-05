package jsonHelpers

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func JSONCompare(a, b []byte) (bool, error) {
	var x, y any

	if err := json.Unmarshal(a, &x); err != nil {
		return false, err
	}

	if err := json.Unmarshal(b, &y); err != nil {
		return false, err
	}

	return reflect.DeepEqual(x, y), nil
}
