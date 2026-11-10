package golden

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestPath_EscapesParameters sends path parameters holding what a path would otherwise read as its own: a slash, a
// percent sign, a question mark. Each must reach the route that takes one parameter, whole.
func TestPath_EscapesParameters(t *testing.T) {
	for _, id := range []string{"x/y", "50%", "a?b#c", "a b"} {
		t.Run(id, func(t *testing.T) {
			var got string

			mux := http.NewServeMux()
			mux.HandleFunc("GET /v1/pages/{id}", func(w http.ResponseWriter, r *http.Request) {
				got = r.PathValue("id")
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
		})
	}
}
