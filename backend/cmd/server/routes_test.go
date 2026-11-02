package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"weatherwatch/internal/database"
	"weatherwatch/internal/forecast"
	"weatherwatch/internal/watchlist"
)

type testWeather struct{}

func (testWeather) SearchPlaces(context.Context, string, int) ([]watchlist.Place, error) {
	return []watchlist.Place{{ID: 658225, Name: "Helsinki"}}, nil
}

func (testWeather) Forecast(context.Context, float64, float64) (forecast.Forecast, error) {
	return forecast.Forecast{Timezone: "Europe/Helsinki", Daily: []forecast.DailyWeather{}}, nil
}

func TestRouterWatchlistFlow(t *testing.T) {
	db, err := database.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := watchlist.NewStore(db)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := newRouter(log,
		watchlist.NewHandler(watchlist.NewService(store, testWeather{})),
		forecast.NewHandler(forecast.NewService(store, testWeather{}, time.Minute)),
	)
	request := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		if rec.Code != status {
			t.Fatalf("%s %s: got %d %s, want %d", method, path, rec.Code, rec.Body, status)
		}
		return rec
	}
	request(http.MethodGet, "/healthz", "", http.StatusOK)
	search := request(http.MethodGet, "/api/places?q=Hel", "", http.StatusOK)
	if !strings.Contains(search.Body.String(), "Helsinki") {
		t.Fatalf("unexpected search response: %s", search.Body)
	}
	request(http.MethodPost, "/api/locations",
		`{"externalId":658225,"name":"Helsinki","latitude":60.17,"longitude":24.94}`,
		http.StatusCreated)
	list := request(http.MethodGet, "/api/locations", "", http.StatusOK)
	var locations []watchlist.Location
	if err := json.Unmarshal(list.Body.Bytes(), &locations); err != nil {
		t.Fatal(err)
	}
	if len(locations) != 1 || locations[0].ID != 1 || locations[0].Name != "Helsinki" {
		t.Fatalf("unexpected locations: %+v", locations)
	}
	weather := request(http.MethodGet, "/api/locations/1/forecast", "", http.StatusOK)
	if !strings.Contains(weather.Body.String(), "Europe/Helsinki") {
		t.Fatalf("unexpected forecast response: %s", weather.Body)
	}
	request(http.MethodDelete, "/api/locations/1", "", http.StatusNoContent)
	request(http.MethodGet, "/api/locations/1/forecast", "", http.StatusNotFound)
	list = request(http.MethodGet, "/api/locations", "", http.StatusOK)
	if strings.TrimSpace(list.Body.String()) != "[]" {
		t.Fatalf("expected empty watchlist, got %s", list.Body)
	}
	request(http.MethodPut, "/api/locations", "", http.StatusMethodNotAllowed)
	request(http.MethodGet, "/missing", "", http.StatusNotFound)
}
