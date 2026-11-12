package httpHelpers

import (
	"context"
	"errors"
	"net/http"
	"revit/internal/jsonHelpers"
)

// DomainError allows domain errors to specify their target HTTP status code.
type DomainError interface {
	error
	StatusCode() int
}

// AppError is a standard implementation of DomainError.
type AppError struct {
	Code    int
	Message string
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) StatusCode() int {
	return e.Code
}

func NewError(code int, message string) AppError {
	return AppError{Code: code, Message: message}
}

// ServiceHandler defines a transport-neutral business logic function.
type ServiceHandler[Req any, Res any] func(ctx context.Context, req Req) (Res, error)

// Adapt converts a transport-neutral ServiceHandler into an http.HandlerFunc.
func Adapt[Req any, Res any](handler ServiceHandler[Req, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		if r.ContentLength > 0 || r.Method == http.MethodPost || r.Method == http.MethodPut {
			var err error
			_, err = RequestJsonToStruct(r, &req)
			if err != nil {
				jsonHelpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
					"success": false,
					"error":   "invalid JSON payload",
				})
				return
			}
		}

		res, err := handler(r.Context(), req)
		if err != nil {
			var domErr DomainError
			if errors.As(err, &domErr) {
				jsonHelpers.WriteJSON(w, domErr.StatusCode(), map[string]any{
					"success": false,
					"error":   domErr.Error(),
				})
				return
			}

			jsonHelpers.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"error":   "internal server error",
			})
			return
		}

		jsonHelpers.WriteJSON(w, http.StatusOK, res)
	}
}
