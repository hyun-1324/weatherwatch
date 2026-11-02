package main

import (
	"log/slog"
	"net/http"

	"weatherwatch/internal/forecast"
	"weatherwatch/internal/httpx"
	"weatherwatch/internal/watchlist"
)

func newRouter(log *slog.Logger, watchlistHandler *watchlist.Handler, forecastHandler *forecast.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/places", watchlistHandler.SearchPlaces)
	mux.HandleFunc("GET /api/locations", watchlistHandler.ListLocations)
	mux.HandleFunc("POST /api/locations", watchlistHandler.CreateLocation)
	mux.HandleFunc("DELETE /api/locations/{id}", watchlistHandler.DeleteLocation)
	mux.HandleFunc("GET /api/locations/{id}/forecast", forecastHandler.LocationForecast)

	return httpx.LogRequests(log, httpx.Recover(log, mux))
}
