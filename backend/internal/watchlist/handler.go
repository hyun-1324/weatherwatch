package watchlist

import (
	"net/http"

	"weatherwatch/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SearchPlaces(w http.ResponseWriter, r *http.Request) {
	places, err := h.svc.SearchPlaces(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, places)
}

func (h *Handler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, locs)
}

// Separate from Location so clients cannot set server-owned fields like id.
type createLocationRequest struct {
	ExternalID int64   `json:"externalId"`
	Name       string  `json:"name"`
	Country    string  `json:"country"`
	Admin1     string  `json:"admin1"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req createLocationRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	loc, err := h.svc.Add(r.Context(), Location{
		ExternalID: req.ExternalID,
		Name:       req.Name,
		Country:    req.Country,
		Admin1:     req.Admin1,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, loc)
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r, "id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	if err := h.svc.Remove(r.Context(), id); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
