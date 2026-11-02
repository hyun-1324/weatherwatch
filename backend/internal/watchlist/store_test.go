package watchlist

import (
	"context"
	"errors"
	"testing"

	"weatherwatch/internal/database"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db)
}

var helsinki = Location{ExternalID: 658225, Name: "Helsinki", Country: "Finland", Latitude: 60.17, Longitude: 24.94}

func TestStoreCreateListDelete(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	created, err := s.Create(ctx, helsinki)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() {
		t.Fatalf("expected ID and CreatedAt to be set: %+v", created)
	}

	got, err := s.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != created {
		t.Fatalf("Get = %+v, want %+v", got, created)
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 location, got %d", len(list))
	}

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, created.ID); !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("Get after delete: err = %v, want ErrLocationNotFound", err)
	}
}

func TestStoreListEmptyIsNotNil(t *testing.T) {
	list, err := newTestStore(t).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if list == nil {
		t.Fatal("want empty slice (encodes as []), got nil (encodes as null)")
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)

	if _, err := s.Create(ctx, helsinki); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, helsinki); !errors.Is(err, ErrLocationAlreadySaved) {
		t.Fatalf("err = %v, want ErrLocationAlreadySaved", err)
	}
}

func TestStoreDeleteMissing(t *testing.T) {
	if err := newTestStore(t).Delete(context.Background(), 42); !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("err = %v, want ErrLocationNotFound", err)
	}
}
