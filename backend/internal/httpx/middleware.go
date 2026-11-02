package httpx

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type requestLog struct {
	mu   sync.Mutex // a handler may record from more than one goroutine
	err  error
	code string
}

type requestLogKey struct{}

// The first error wins: it is usually the cause.
func recordError(r *http.Request, err error, code string) {
	rl, ok := r.Context().Value(requestLogKey{}).(*requestLog)
	if !ok {
		// No LogRequests middleware (e.g. in a unit test): log directly.
		slog.Error("request error", "path", r.URL.Path, "error", err, "error_code", code)
		return
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.err == nil {
		rl.err, rl.code = err, code
	}
}

func LogRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rl := &requestLog{}
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), requestLogKey{}, rl)))

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.Status(),
			"duration", time.Since(start),
		}
		rl.mu.Lock()
		if rl.err != nil {
			attrs = append(attrs, "error", rl.err.Error(), "error_code", rl.code)
		}
		rl.mu.Unlock()

		log.Log(r.Context(), levelFor(rec.Status()), "request", attrs...)
	})
}

func levelFor(status int) slog.Level {
	switch {
	case status == http.StatusBadGateway:
		return slog.LevelWarn // a dependency failed, not our code
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Recover(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler { //nolint:errorlint // net/http compares this exact value
				panic(v) // deliberate abort; let net/http handle it
			}
			err := fmt.Errorf("panic: %v", v)
			recordError(r, err, "INTERNAL")
			log.Error("panic recovered", "path", r.URL.Path, "error", err, "stack", string(debug.Stack()))

			// If the handler already started the response we cannot change it.
			if !rec.wroteHeader {
				writeErrorBody(rec, http.StatusInternalServerError, "INTERNAL", "Internal server error")
			}
		}()
		next.ServeHTTP(rec, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rec *statusRecorder) WriteHeader(code int) {
	if !rec.wroteHeader {
		rec.status, rec.wroteHeader = code, true
	}
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.wroteHeader {
		rec.WriteHeader(http.StatusOK)
	}
	return rec.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (rec *statusRecorder) Unwrap() http.ResponseWriter { return rec.ResponseWriter }

func (rec *statusRecorder) Status() int {
	if !rec.wroteHeader {
		return http.StatusOK
	}
	return rec.status
}
