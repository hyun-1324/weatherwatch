package forecast

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"weatherwatch/internal/watchlist"
)

// Fakes shared by the service and handler tests in this package.

type fakeLocations map[int64]watchlist.Location

func (f fakeLocations) Get(_ context.Context, id int64) (watchlist.Location, error) {
	loc, ok := f[id]
	if !ok {
		return watchlist.Location{}, watchlist.ErrLocationNotFound
	}
	return loc, nil
}

type fakeProvider struct {
	forecast Forecast
	err      error
	calls    int
}

func (f *fakeProvider) Forecast(context.Context, float64, float64) (Forecast, error) {
	f.calls++
	return f.forecast, f.err
}

func TestForecastIsCachedUntilTTL(t *testing.T) {
	ctx := context.Background()
	provider := &fakeProvider{forecast: Forecast{Timezone: "Europe/Helsinki"}}
	svc := NewService(fakeLocations{1: {ID: 1}}, provider, 10*time.Minute)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	for range 3 {
		if _, err := svc.ForLocation(ctx, 1); err != nil {
			t.Fatal(err)
		}
	}
	if provider.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1 (cached)", provider.calls)
	}

	now = now.Add(11 * time.Minute)
	if _, err := svc.ForLocation(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 2 {
		t.Fatalf("upstream calls = %d, want 2 after TTL expired", provider.calls)
	}
}

func TestForecastUnknownLocation(t *testing.T) {
	provider := &fakeProvider{}
	svc := NewService(fakeLocations{}, provider, time.Minute)

	if _, err := svc.ForLocation(context.Background(), 99); !errors.Is(err, watchlist.ErrLocationNotFound) {
		t.Fatalf("err = %v, want ErrLocationNotFound", err)
	}
	if provider.calls != 0 {
		t.Fatal("upstream should not be called for an unknown location")
	}
}

func TestForecastErrorsAreNotCached(t *testing.T) {
	ctx := context.Background()
	provider := &fakeProvider{err: errors.New("503")}
	svc := NewService(fakeLocations{1: {ID: 1}}, provider, time.Minute)

	if _, err := svc.ForLocation(ctx, 1); !errors.Is(err, ErrForecastUnavailable) {
		t.Fatalf("err = %v, want ErrForecastUnavailable", err)
	}

	provider.err = nil
	if _, err := svc.ForLocation(ctx, 1); err != nil {
		t.Fatalf("second call should retry upstream, got %v", err)
	}
	if provider.calls != 2 {
		t.Fatalf("upstream calls = %d, want 2", provider.calls)
	}
}

type blockingProvider struct {
	release chan struct{}
	calls   atomic.Int32
}

func (p *blockingProvider) Forecast(context.Context, float64, float64) (Forecast, error) {
	p.calls.Add(1)
	<-p.release
	return Forecast{Timezone: "Europe/Helsinki"}, nil
}

func TestConcurrentMissesShareOneUpstreamCall(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		provider := &blockingProvider{release: make(chan struct{})}
		svc := NewService(fakeLocations{1: {ID: 1}}, provider, time.Minute)

		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				f, err := svc.ForLocation(context.Background(), 1)
				if err != nil || f.Timezone != "Europe/Helsinki" {
					t.Errorf("got %+v, %v", f, err)
				}
			})
		}
		synctest.Wait() // all 10 are now blocked: one in the provider, nine waiting on it
		close(provider.release)
		wg.Wait()

		if n := provider.calls.Load(); n != 1 {
			t.Fatalf("upstream calls = %d, want 1", n)
		}
	})
}
