package forecast

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serve(t *testing.T, provider *fakeProvider, id string) *httptest.ResponseRecorder {
	t.Helper()
	svc := NewService(fakeLocations{1: {ID: 1}}, provider, time.Minute)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	NewHandler(svc).LocationForecast(rec, r)
	return rec
}

func TestHandlerForecast(t *testing.T) {
	rec := serve(t, &fakeProvider{forecast: Forecast{Timezone: "Europe/Helsinki"}}, "1")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Europe/Helsinki") {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body)
	}
}

func TestHandlerForecastErrors(t *testing.T) {
	cases := []struct {
		name       string
		provider   *fakeProvider
		id         string
		wantStatus int
		wantCode   string
	}{
		{"invalid id", &fakeProvider{}, "abc", http.StatusBadRequest, "INVALID_ID"},
		{"unknown location", &fakeProvider{}, "99", http.StatusNotFound, "LOCATION_NOT_FOUND"},
		{"provider down", &fakeProvider{err: errors.New("503")}, "1", http.StatusBadGateway, "FORECAST_UNAVAILABLE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(t, tc.provider, tc.id)
			if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), `"code":"`+tc.wantCode+`"`) {
				t.Fatalf("got %d %s, want %d %s", rec.Code, rec.Body, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
