package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"weatherwatch/internal/apperr"
)

var (
	ErrInvalidBody = apperr.Invalid("INVALID_BODY", "Invalid JSON body")
	ErrInvalidID   = apperr.Invalid("INVALID_ID", "id must be a positive integer")
)

type errorResponse struct {
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// nginx's non-standard code for a request the client abandoned.
const statusClientClosedRequest = 499

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	_, tooLarge := errors.AsType[*http.MaxBytesError](err)
	appErr, isAppErr := errors.AsType[*apperr.Error](err)

	switch {
	case r.Context().Err() != nil:
		// The client went away; this is not a server fault.
		recordError(r, err, "CLIENT_CLOSED_REQUEST")
		writeErrorBody(w, statusClientClosedRequest, "CLIENT_CLOSED_REQUEST", "Client closed request")
	case tooLarge:
		recordError(r, err, "BODY_TOO_LARGE")
		writeErrorBody(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "Request body too large")
	case isAppErr:
		recordError(r, err, appErr.Code)
		writeErrorBody(w, appErr.Status, appErr.Code, appErr.Message)
	default:
		recordError(r, err, "INTERNAL")
		writeErrorBody(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
	}
}

func writeErrorBody(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorResponse{StatusCode: status, Code: code, Message: message})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBody, err)
	}
	return nil
}

func PathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidID
	}
	return id, nil
}
