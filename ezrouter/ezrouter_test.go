package ezrouter

import (
	"net/http"
	"testing"
)

func TestPlaceholderRoute_Match(t *testing.T) {
	route := PlaceholderRoute{
		Method:       http.MethodGet,
		Pattern:      "/catalog/artist/:uri",
	}
	if err := route.compile(); err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodGet, "/catalog/artist/test-artist", nil)
	matches := route.Match(req)
	if matches == nil {
		t.Errorf("Failed to match %s against %s", req.URL.Path, route.Pattern)
	}
}
