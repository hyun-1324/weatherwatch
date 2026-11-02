package forecast

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"weatherwatch/internal/watchlist"
)

type Provider interface {
	Forecast(ctx context.Context, lat, lon float64) (Forecast, error)
}

type LocationGetter interface {
	Get(ctx context.Context, id int64) (watchlist.Location, error)
}

type Service struct {
	locations LocationGetter
	provider  Provider
	ttl       time.Duration
	now       func() time.Time

	inflight singleflight.Group

	mu    sync.Mutex
	cache map[int64]cachedForecast
}

type cachedForecast struct {
	forecast  Forecast
	expiresAt time.Time
}

func NewService(locations LocationGetter, provider Provider, ttl time.Duration) *Service {
	return &Service{
		locations: locations,
		provider:  provider,
		ttl:       ttl,
		now:       time.Now,
		cache:     make(map[int64]cachedForecast),
	}
}

func (s *Service) ForLocation(ctx context.Context, locationID int64) (Forecast, error) {
	loc, err := s.locations.Get(ctx, locationID)
	if err != nil {
		return Forecast{}, err
	}

	if f, ok := s.cached(locationID); ok {
		return f, nil
	}

	v, err, _ := s.inflight.Do(strconv.FormatInt(locationID, 10), func() (any, error) {
		// A call that finished just before this one may have filled the cache.
		if f, ok := s.cached(locationID); ok {
			return f, nil
		}
		// Detached so other callers waiting on this result don't fail if this one disconnects.
		f, err := s.provider.Forecast(context.WithoutCancel(ctx), loc.Latitude, loc.Longitude)
		if err != nil {
			return nil, err
		}
		s.mu.Lock()
		s.cache[locationID] = cachedForecast{forecast: f, expiresAt: s.now().Add(s.ttl)}
		s.mu.Unlock()
		return f, nil
	})
	if err != nil {
		return Forecast{}, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}
	return v.(Forecast), nil
}

func (s *Service) cached(locationID int64) (Forecast, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[locationID]
	if !ok || s.now().After(entry.expiresAt) {
		return Forecast{}, false
	}
	return entry.forecast, true
}
