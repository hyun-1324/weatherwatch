package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"weatherwatch/internal/apperr"
)

var errThingNotFound = apperr.NotFound("THING_NOT_FOUND", "Thing not found")

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body
}

func TestWriteErrorUsesStatusAndCode(t *testing.T) {
	cases := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{apperr.Invalid("NAME_REQUIRED", "Name is required"), 400, "NAME_REQUIRED"},
		{errThingNotFound, 404, "THING_NOT_FOUND"},
		{fmt.Errorf("load 3: %w", errThingNotFound), 404, "THING_NOT_FOUND"},
		{apperr.Conflict("ALREADY_SAVED", "Already saved"), 409, "ALREADY_SAVED"},
		{fmt.Errorf("%w: %w", apperr.Upstream("DOWN", "Down"), errors.New("timeout")), 502, "DOWN"},
		{errors.New("disk full"), 500, "INTERNAL"},
		{apperr.New(http.StatusTooManyRequests, "RATE_LIMITED", "Request limit exceeded"), 429, "RATE_LIMITED"},
	}
	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), tc.err)

			body := decodeError(t, rec)
			if rec.Code != tc.wantStatus || body.StatusCode != tc.wantStatus || body.Code != tc.wantCode {
				t.Fatalf("got %d %+v, want %d %s", rec.Code, body, tc.wantStatus, tc.wantCode)
			}
		})
	}
}

func TestWriteErrorDoesNotLeakInternalErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errors.New("secret db path /var/x"))
	if strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("internal error leaked to client: %s", rec.Body)
	}
}

func TestWriteErrorClientDisconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	rec := httptest.NewRecorder()
	WriteError(rec, req, context.Canceled)
	if rec.Code != statusClientClosedRequest {
		t.Fatalf("status = %d, want 499", rec.Code)
	}
}

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"malformed", `{`, http.StatusBadRequest},
		{"unknown field", `{"name":"x","extra":1}`, http.StatusBadRequest},
		{"too large", `{"name":"` + strings.Repeat("x", 1<<17) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			var p payload
			err := DecodeJSON(rec, req, &p)
			if err == nil {
				t.Fatal("expected an error")
			}
			WriteError(rec, req, err)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestPathID(t *testing.T) {
	cases := map[string]bool{"1": true, "42": true, "0": false, "-1": false, "abc": false}
	for raw, wantOK := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetPathValue("id", raw)
		_, err := PathID(r, "id")
		if (err == nil) != wantOK {
			t.Errorf("PathID(%q): err = %v, want ok=%v", raw, err, wantOK)
		}
	}
}

// --- middleware ---

func newJSONLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("bad log line %q: %v", line, err)
		}
		lines = append(lines, m)
	}
	return lines
}

func TestLogRequestsWritesOneLineWithError(t *testing.T) {
	var buf bytes.Buffer
	h := LogRequests(newJSONLogger(&buf), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, fmt.Errorf("load 3: %w", errThingNotFound))
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/things/3", nil))

	lines := logLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("want exactly 1 log line, got %d: %s", len(lines), buf.String())
	}
	l := lines[0]
	if l["status"] != float64(404) || l["error"] != "load 3: THING_NOT_FOUND: Thing not found" || l["error_code"] != "THING_NOT_FOUND" || l["level"] != "INFO" {
		t.Fatalf("unexpected log line: %v", l)
	}
}

func TestLogRequestsLevels(t *testing.T) {
	cases := map[int]string{200: "INFO", 404: "INFO", 500: "ERROR", 502: "WARN", 499: "INFO"}
	for status, wantLevel := range cases {
		var buf bytes.Buffer
		h := LogRequests(newJSONLogger(&buf), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		if got := logLines(t, &buf)[0]["level"]; got != wantLevel {
			t.Errorf("status %d: level = %v, want %s", status, got, wantLevel)
		}
	}
}

func TestRecover(t *testing.T) {
	var buf bytes.Buffer
	log := newJSONLogger(&buf)
	h := LogRequests(log, Recover(log, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("nil map")
	})))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError || decodeError(t, rec).Code != "INTERNAL" {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body)
	}
	lines := logLines(t, &buf)
	last := lines[len(lines)-1]
	if last["status"] != float64(500) || last["error_code"] != "INTERNAL" || last["error"] != "panic: nil map" || last["level"] != "ERROR" {
		t.Fatalf("request log should record the panic: %v", last)
	}
}

func TestWriteErrorKeepsWrappedCauseOutOfResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	appErr := apperr.Upstream("FORECAST_UNAVAILABLE", "The weather provider could not supply the forecast")
	err := fmt.Errorf("%w: %w", appErr, errors.New("private upstream diagnostic"))
	WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), err)
	body := decodeError(t, rec)
	if body.Message != appErr.Message || body.Code != appErr.Code || rec.Code != http.StatusBadGateway {
		t.Fatalf("unexpected error response: %d %+v", rec.Code, body)
	}
	if strings.Contains(rec.Body.String(), "private upstream diagnostic") {
		t.Fatal("wrapped cause leaked to API response")
	}
}
