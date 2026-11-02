package watchlist

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type request struct {
	handle func(*Handler, http.ResponseWriter, *http.Request)
	id     string // the {id} path value, if the handler reads one
	query  string
	body   string
}

func serve(t *testing.T, repo *fakeRepo, req request) *httptest.ResponseRecorder {
	t.Helper()
	places := &fakePlaces{places: []Place{{ID: 1, Name: "Helsinki"}}}
	handler := NewHandler(NewService(repo, places))

	r := httptest.NewRequest(http.MethodGet, "/?"+req.query, strings.NewReader(req.body))
	if req.id != "" {
		r.SetPathValue("id", req.id)
	}
	rec := httptest.NewRecorder()
	req.handle(handler, rec, r)
	return rec
}

const helsinkiJSON = `{"externalId":658225,"name":"Helsinki","country":"Finland","admin1":"Uusimaa","latitude":60.17,"longitude":24.94}`

func TestHandlerSuccess(t *testing.T) {
	cases := map[string]struct {
		req        request
		wantStatus int
		wantBody   string
	}{
		"search": {request{handle: (*Handler).SearchPlaces, query: "q=Hel"}, http.StatusOK, `"Helsinki"`},
		"list":   {request{handle: (*Handler).ListLocations}, http.StatusOK, `[]`},
		"create": {request{handle: (*Handler).CreateLocation, body: helsinkiJSON}, http.StatusCreated, `"id":1`},
		"delete": {request{handle: (*Handler).DeleteLocation, id: "1"}, http.StatusNoContent, ``},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, &fakeRepo{locations: []Location{}}, tc.req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.wantStatus, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("body %s does not contain %s", rec.Body, tc.wantBody)
			}
		})
	}
}

func TestCreateMapsRequestToLocation(t *testing.T) {
	repo := &fakeRepo{}
	serve(t, repo, request{handle: (*Handler).CreateLocation, body: helsinkiJSON})

	want := Location{ExternalID: 658225, Name: "Helsinki", Country: "Finland", Admin1: "Uusimaa", Latitude: 60.17, Longitude: 24.94}
	if len(repo.created) != 1 || repo.created[0] != want {
		t.Fatalf("repository got %+v, want %+v", repo.created, want)
	}
}

func TestHandlerErrors(t *testing.T) {
	create := func(body string) request { return request{handle: (*Handler).CreateLocation, body: body} }
	remove := func(id string) request { return request{handle: (*Handler).DeleteLocation, id: id} }
	cases := map[string]struct {
		repo       *fakeRepo
		req        request
		wantStatus int
		wantCode   string
	}{
		"short query":    {&fakeRepo{}, request{handle: (*Handler).SearchPlaces, query: "q=H"}, http.StatusBadRequest, "QUERY_TOO_SHORT"},
		"malformed json": {&fakeRepo{}, create(`{`), http.StatusBadRequest, "INVALID_BODY"},
		"unknown field":  {&fakeRepo{}, create(`{"externalId":1,"id":5}`), http.StatusBadRequest, "INVALID_BODY"},
		"invalid values": {&fakeRepo{}, create(`{"externalId":1,"name":"X","latitude":91}`), http.StatusBadRequest, "INVALID_LATITUDE"},
		"non-numeric id": {&fakeRepo{}, remove("abc"), http.StatusBadRequest, "INVALID_ID"},
		"already saved":  {&fakeRepo{err: ErrLocationAlreadySaved}, create(helsinkiJSON), http.StatusConflict, "LOCATION_ALREADY_SAVED"},
		"missing":        {&fakeRepo{err: ErrLocationNotFound}, remove("9"), http.StatusNotFound, "LOCATION_NOT_FOUND"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, tc.repo, tc.req)
			if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), `"code":"`+tc.wantCode+`"`) {
				t.Fatalf("got %d %s, want %d %s", rec.Code, rec.Body, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
