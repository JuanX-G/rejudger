package httpHelpers

import (
	"encoding/json"
	"net/http"
)

func RequestJsonToStruct[T any](r *http.Request, v T) (T, error) {
	var empty T

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return empty, err
	}
	return v, nil
}

type RouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

func CapInt(i, cap int) int {
	if i > cap {
		return cap
	} else {
		return i
	}
}
