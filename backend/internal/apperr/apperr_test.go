package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

var errThingNotFound = NotFound("THING_NOT_FOUND", "Thing not found")

func TestWrappedSentinelIsStillRecognised(t *testing.T) {
	err := fmt.Errorf("load thing 3: %w", errThingNotFound)

	if !errors.Is(err, errThingNotFound) {
		t.Fatal("errors.Is should see through %w wrapping")
	}
	got, ok := errors.AsType[*Error](err)
	if !ok || got.Code != "THING_NOT_FOUND" || got.Status != http.StatusNotFound {
		t.Fatalf("AsType = %+v, want the 404 sentinel", got)
	}
}

func TestSentinelWrappingACause(t *testing.T) {
	errDown := Upstream("THING_UNAVAILABLE", "Thing service is unavailable")
	cause := errors.New("connection refused")
	err := fmt.Errorf("%w: %w", errDown, cause)

	if !errors.Is(err, errDown) || !errors.Is(err, cause) {
		t.Fatal("both the sentinel and the cause should be in the chain")
	}
}
