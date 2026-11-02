package watchlist

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"weatherwatch/internal/apperr"
)

// Fakes shared by the service and handler tests in this package.

type fakeRepo struct {
	locations []Location
	created   []Location
	err       error
}

func (r *fakeRepo) List(context.Context) ([]Location, error) { return r.locations, r.err }

func (r *fakeRepo) Create(_ context.Context, loc Location) (Location, error) {
	if r.err != nil {
		return Location{}, r.err
	}
	r.created = append(r.created, loc)
	loc.ID = int64(len(r.created))
	return loc, nil
}

func (r *fakeRepo) Delete(context.Context, int64) error { return r.err }

type fakePlaces struct {
	places []Place
	err    error
	calls  int
}

func (f *fakePlaces) SearchPlaces(context.Context, string, int) ([]Place, error) {
	f.calls++
	return f.places, f.err
}

func TestSearchPlacesValidatesQuery(t *testing.T) {
	places := &fakePlaces{}
	svc := NewService(&fakeRepo{}, places)

	for _, q := range []string{"  a ", "서"} { // "서" is one character but three bytes
		if _, err := svc.SearchPlaces(context.Background(), q); !errors.Is(err, ErrQueryTooShort) {
			t.Fatalf("SearchPlaces(%q): err = %v, want ErrQueryTooShort", q, err)
		}
	}
	if places.calls != 0 {
		t.Fatal("upstream should not be called for an invalid query")
	}
}

func TestSearchPlacesWrapsUpstreamErrors(t *testing.T) {
	svc := NewService(&fakeRepo{}, &fakePlaces{err: errors.New("timeout")})

	if _, err := svc.SearchPlaces(context.Background(), "Helsinki"); !errors.Is(err, ErrPlaceSearchUnavailable) {
		t.Fatalf("err = %v, want ErrPlaceSearchUnavailable", err)
	}
}

func TestAddRejectsInvalidLocation(t *testing.T) {
	repo := &fakeRepo{}
	invalid := Location{ExternalID: 1, Name: "Helsinki", Latitude: 91}

	_, err := NewService(repo, &fakePlaces{}).Add(context.Background(), invalid)
	if e, ok := errors.AsType[*apperr.Error](err); !ok || e.Status != http.StatusBadRequest {
		t.Fatalf("err = %v, want an invalid-input error", err)
	}
	if len(repo.created) != 0 {
		t.Fatal("invalid location must not reach the repository")
	}
}

func TestAddTrimsName(t *testing.T) {
	repo := &fakeRepo{}
	loc := Location{ExternalID: 1, Name: "  Helsinki ", Latitude: 60, Longitude: 24}

	if _, err := NewService(repo, &fakePlaces{}).Add(context.Background(), loc); err != nil {
		t.Fatal(err)
	}
	if got := repo.created[0].Name; got != "Helsinki" {
		t.Fatalf("name = %q, want trimmed", got)
	}
}
