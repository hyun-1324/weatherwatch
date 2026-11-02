package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New()
	c.GeocodingURL = srv.URL
	c.ForecastURL = srv.URL
	return c
}

func TestSearchPlaces(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("name"); got != "Helsinki" {
			t.Errorf("name param = %q, want Helsinki", got)
		}
		w.Write([]byte(`{"results":[{"id":658225,"name":"Helsinki","latitude":60.17,"longitude":24.94,
			"country":"Finland","country_code":"FI","admin1":"Uusimaa","timezone":"Europe/Helsinki"}]}`))
	})

	places, err := c.SearchPlaces(context.Background(), "Helsinki", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(places) != 1 || places[0].ID != 658225 || places[0].Country != "Finland" {
		t.Fatalf("unexpected places: %+v", places)
	}
}

func TestSearchPlacesNoResults(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"generationtime_ms":0.1}`))
	})

	places, err := c.SearchPlaces(context.Background(), "zzzz", 5)
	if err != nil {
		t.Fatal(err)
	}
	if places == nil || len(places) != 0 {
		t.Fatalf("want empty non-nil slice (encodes as [] not null), got %#v", places)
	}
}

func TestForecast(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"timezone":"Europe/Helsinki",
			"current":{"time":"2026-10-28T08:45","temperature_2m":10.4,"apparent_temperature":5.3,
				"relative_humidity_2m":85,"wind_speed_10m":31.3,"weather_code":3,"is_day":1},
			"daily":{"time":["2026-10-28","2026-10-29"],"weather_code":[51,0],
				"temperature_2m_max":[10.9,11.4],"temperature_2m_min":[10.0,9.9],
				"precipitation_probability_max":[49,null]}}`))
	})

	f, err := c.Forecast(context.Background(), 60.17, 24.94)
	if err != nil {
		t.Fatal(err)
	}
	if f.Current.Condition != "Overcast" || !f.Current.IsDay || f.Current.Humidity != 85 {
		t.Errorf("unexpected current: %+v", f.Current)
	}
	if len(f.Daily) != 2 {
		t.Fatalf("want 2 days, got %d", len(f.Daily))
	}
	if p := f.Daily[0].PrecipitationProbability; p == nil || *p != 49 {
		t.Errorf("day 0 precipitation = %v, want 49", p)
	}
	if f.Daily[1].PrecipitationProbability != nil {
		t.Errorf("day 1 precipitation should be nil for null upstream value")
	}
	if f.Daily[1].Condition != "Clear sky" {
		t.Errorf("day 1 condition = %q", f.Daily[1].Condition)
	}
}

func TestForecastRejectsMismatchedSeries(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"daily":{"time":["2026-10-28","2026-10-29"],"weather_code":[1],
			"temperature_2m_max":[1,2],"temperature_2m_min":[1,2],"precipitation_probability_max":[1,2]}}`))
	})

	if _, err := c.Forecast(context.Background(), 0, 0); err == nil {
		t.Fatal("expected error for mismatched daily series")
	}
}

func TestUpstreamErrorStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":true,"reason":"bad"}`, http.StatusBadRequest)
	})

	if _, err := c.Forecast(context.Background(), 0, 0); err == nil {
		t.Fatal("expected error for non-200 upstream response")
	}
}
