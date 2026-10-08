package golden

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
)

// queryService records the arguments QueryStyles was called with; it implements nothing else.
type queryService struct {
	Service
	page int
	got  QueryStylesParams
}

func (s *queryService) QueryStyles(_ context.Context, page int, params QueryStylesParams) error {
	s.page, s.got = page, params
	return nil
}

// TestQuery_RoundTrip sends a value of every query parameter style and format, each with the characters that would
// break it if escaped wrongly, and the server must read what the client wrote.
func TestQuery_RoundTrip(t *testing.T) {
	svc := &queryService{}
	mux := http.NewServeMux()
	RegisterService(svc, mux, "/v1")

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	base, err := url.Parse(srv.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(WithBaseURL(base))
	if err != nil {
		t.Fatal(err)
	}

	want := QueryStylesParams{
		Csv:        []int{1, 2},
		Spaced:     []string{"a b", "c"},
		Piped:      []string{"x|y", "z"},
		Color:      Color{R: 1, Name: "x"},
		ColorCsv:   Color{R: 3, G: new(0), Name: "a,b"},
		ColorPiped: Color{R: 4, Name: "p|q"},
		ColorDeep:  Color{R: 2, G: new(5), Fade: new(30 * time.Second)},
		Labels:     map[string]string{"a b": "c&d"},
		Weights:    map[string]int{"w": 1, "v": 2},
		Extra:      map[string]string{"other": "y"},
		Next:       "https://x.example/a?b=c&d+e#f",
		Paths:      []string{"/a/b", "c:d"},
		Cells:      []string{"x/y", "z,w"},
		Every:      90 * time.Second,
		At:         time.Unix(1791453600, 0),
		On:         civil.Date{Year: 2026, Month: 10, Day: 8},
		Site:       url.URL{Scheme: "https", Host: "x.example", Path: "/p"},
		IP:         net.ParseIP("10.0.0.1"),
		Level:      2,
		Ratio:      0.5,
	}

	if err := c.QueryStyles(t.Context(), 3, want); err != nil {
		t.Fatal(err)
	}

	if svc.page != 3 || !reflect.DeepEqual(svc.got, want) {
		t.Fatalf("got page %d and %+v\nwant page 3 and %+v", svc.page, svc.got, want)
	}

	// an object not set sends nothing, though it requires a member
	if err := c.QueryStyles(t.Context(), 3, QueryStylesParams{}); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(svc.got, QueryStylesParams{}) {
		t.Fatalf("got %+v, want nothing set", svc.got)
	}
}
