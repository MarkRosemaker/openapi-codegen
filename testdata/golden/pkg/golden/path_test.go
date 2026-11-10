package golden

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestPath_EscapesParameters sends path parameters holding what a path would otherwise read as its own: a percent
// sign, a question mark, a space. Each must reach the route whole, sent as the path wants it. A slash stays a slash,
// as in a package path such as golang.org/x/time/rate, and a literal %2F stays that too.
func TestPath_EscapesParameters(t *testing.T) {
	for id, want := range map[string]string{
		"golang.org/x/time/rate": "/v1/pages/golang.org/x/time/rate",
		"50%":                    "/v1/pages/50%25",
		"a?b#c":                  "/v1/pages/a%3Fb%23c",
		"a b":                    "/v1/pages/a%20b",
		"x%2Fy":                  "/v1/pages/x%252Fy",
	} {
		t.Run(id, func(t *testing.T) {
			var got, sent string

			mux := http.NewServeMux()
			mux.HandleFunc("GET /v1/pages/{id...}", func(w http.ResponseWriter, r *http.Request) {
				got, sent = r.PathValue("id"), r.URL.EscapedPath()
				w.WriteHeader(http.StatusTeapot)
			})

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

			if _, err := c.GetPage(t.Context(), id); err == nil {
				t.Fatal("expected the teapot to fail the call")
			}

			if got != id {
				t.Errorf("the route got %q, want %q", got, id)
			}

			if sent != want {
				t.Errorf("sent to %s, want %s", sent, want)
			}
		})
	}
}
