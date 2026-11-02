package watchlist

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Repository interface {
	List(ctx context.Context) ([]Location, error)
	Create(ctx context.Context, loc Location) (Location, error)
	Delete(ctx context.Context, id int64) error
}

type PlaceSearcher interface {
	SearchPlaces(ctx context.Context, query string, limit int) ([]Place, error)
}

const (
	minQueryLength   = 2
	maxSearchResults = 5
)

type Service struct {
	repo   Repository
	places PlaceSearcher
}

func NewService(repo Repository, places PlaceSearcher) *Service {
	return &Service{repo: repo, places: places}
}

func (s *Service) SearchPlaces(ctx context.Context, query string) ([]Place, error) {
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < minQueryLength {
		return nil, ErrQueryTooShort
	}
	places, err := s.places.SearchPlaces(ctx, query, maxSearchResults)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPlaceSearchUnavailable, err)
	}
	return places, nil
}

func (s *Service) List(ctx context.Context) ([]Location, error) {
	return s.repo.List(ctx)
}

// Duplicates are rejected by the UNIQUE constraint; checking here would race.
func (s *Service) Add(ctx context.Context, loc Location) (Location, error) {
	loc.Name = strings.TrimSpace(loc.Name)
	if err := loc.Validate(); err != nil {
		return Location{}, err
	}
	return s.repo.Create(ctx, loc)
}

func (s *Service) Remove(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
