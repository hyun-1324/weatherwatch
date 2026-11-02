package watchlist

import (
	"math"
	"strings"
	"time"

	"weatherwatch/internal/apperr"
)

type Place struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Codes are part of the API contract.
var (
	ErrQueryTooShort          = apperr.Invalid("QUERY_TOO_SHORT", "Search query must be at least 2 characters")
	ErrLocationNotFound       = apperr.NotFound("LOCATION_NOT_FOUND", "Location not found")
	ErrLocationAlreadySaved   = apperr.Conflict("LOCATION_ALREADY_SAVED", "A location with this externalId already exists")
	ErrPlaceSearchUnavailable = apperr.Upstream("PLACE_SEARCH_UNAVAILABLE", "The geocoding provider could not complete the search")

	ErrInvalidExternalID = apperr.Invalid("INVALID_EXTERNAL_ID", "externalId must be a positive integer")
	ErrNameRequired      = apperr.Invalid("NAME_REQUIRED", "name is required")
	ErrInvalidLatitude   = apperr.Invalid("INVALID_LATITUDE", "latitude must be between -90 and 90")
	ErrInvalidLongitude  = apperr.Invalid("INVALID_LONGITUDE", "longitude must be between -180 and 180")
)

type Location struct {
	ID         int64     `json:"id"`
	ExternalID int64     `json:"externalId"` // Place.ID from the geocoding provider
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	Admin1     string    `json:"admin1"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (l Location) Validate() error {
	switch {
	case l.ExternalID <= 0:
		return ErrInvalidExternalID
	case strings.TrimSpace(l.Name) == "":
		return ErrNameRequired
	case math.IsNaN(l.Latitude) || math.Abs(l.Latitude) > 90:
		return ErrInvalidLatitude
	case math.IsNaN(l.Longitude) || math.Abs(l.Longitude) > 180:
		return ErrInvalidLongitude
	}
	return nil
}
