package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/onezerotoys/todo-app/backend/internal/todo"
)

// ErrorEnvelope is the shape returned for every non-2xx response.
//
//	{"error":{"code":"...","message":"...","details":[{"field":"...","message":"..."}]}}
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the inner part of the error envelope.
type ErrorBody struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []todo.FieldError  `json:"details,omitempty"`
}

// WriteJSON serialises v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a normalised error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string, details ...todo.FieldError) {
	body := ErrorEnvelope{Error: ErrorBody{Code: code, Message: message, Details: details}}
	if body.Error.Details == nil {
		body.Error.Details = []todo.FieldError{}
	}
	WriteJSON(w, status, body)
}

// WriteInternal logs the underlying error and returns a generic 500 to the client.
func WriteInternal(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}

// DecodeJSON parses a JSON body into v. On failure it writes a 400 envelope
// and returns false so the caller can bail out.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Body == nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "request body is empty")
		return false
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "invalid JSON: "+err.Error())
		return false
	}
	return true
}
