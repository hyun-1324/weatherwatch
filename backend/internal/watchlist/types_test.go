package watchlist

import (
	"errors"
	"math"
	"testing"
)

func TestLocationValidate(t *testing.T) {
	valid := Location{ExternalID: 1, Name: "Helsinki", Latitude: 60.17, Longitude: 24.94}

	if err := valid.Validate(); err != nil {
		t.Fatalf("valid location: unexpected error %v", err)
	}

	cases := map[string]struct {
		mutate func(*Location)
		want   error
	}{
		"missing external id": {func(l *Location) { l.ExternalID = 0 }, ErrInvalidExternalID},
		"empty name":          {func(l *Location) { l.Name = "" }, ErrNameRequired},
		"blank name":          {func(l *Location) { l.Name = "   " }, ErrNameRequired},
		"latitude too big":    {func(l *Location) { l.Latitude = 90.1 }, ErrInvalidLatitude},
		"latitude too small":  {func(l *Location) { l.Latitude = -91 }, ErrInvalidLatitude},
		"longitude too big":   {func(l *Location) { l.Longitude = 181 }, ErrInvalidLongitude},
		"longitude too small": {func(l *Location) { l.Longitude = -180.5 }, ErrInvalidLongitude},
		"latitude NaN":        {func(l *Location) { l.Latitude = math.NaN() }, ErrInvalidLatitude},
		"longitude NaN":       {func(l *Location) { l.Longitude = math.NaN() }, ErrInvalidLongitude},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			loc := valid
			tc.mutate(&loc)

			if err := loc.Validate(); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestLocationValidateAcceptsBoundaries(t *testing.T) {
	for _, loc := range []Location{
		{ExternalID: 1, Name: "North Pole", Latitude: 90, Longitude: 180},
		{ExternalID: 1, Name: "South Pole", Latitude: -90, Longitude: -180},
	} {
		if err := loc.Validate(); err != nil {
			t.Errorf("%s: unexpected error %v", loc.Name, err)
		}
	}
}
