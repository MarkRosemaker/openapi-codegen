package golden

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// sentPath calls the client against a server that records the path it was sent, as it was sent.
func sentPath(t *testing.T, call func(c *Client) error) string {
	t.Helper()

	var sent string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent = r.URL.EscapedPath()
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(srv.Close)

	base, err := url.Parse(srv.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(WithBaseURL(base))
	if err != nil {
		t.Fatal(err)
	}

	if err := call(c); err == nil {
		t.Fatal("expected the teapot to fail the call")
	}

	return sent
}

// TestPath_EscapesParameters sends path parameters holding what a path would otherwise read as its own. Each is sent
// as one segment: a percent sign would make the client drop the path, and a slash, which OpenAPI says a path
// parameter does not hold, would add segments, as ../admin would step out of the endpoint.
func TestPath_EscapesParameters(t *testing.T) {
	for id, want := range map[string]string{
		"x/y":      "/v1/pages/x%2Fy",
		"50%":      "/v1/pages/50%25",
		"a?b#c":    "/v1/pages/a%3Fb%23c",
		"a b":      "/v1/pages/a%20b",
		"../admin": "/v1/pages/..%2Fadmin",
	} {
		t.Run(id, func(t *testing.T) {
			if got := sentPath(t, func(c *Client) error { _, err := c.GetPage(t.Context(), id); return err }); got != want {
				t.Errorf("sent to %s, want %s", got, want)
			}
		})
	}
}

// fileService records the path GetFile was called with; it implements nothing else.
type fileService struct {
	Service
	got string
}

func (s *fileService) GetFile(_ context.Context, path string) error {
	s.got = path
	return nil
}

// TestPath_KeepsSlashes sends a path parameter marked x-allow-slashes, a path of its own: its slashes stay, what lies
// between them is escaped, and a part of . or .. is escaped too, so the client keeps it in the path. The generated
// server must hand the service what the client was given.
func TestPath_KeepsSlashes(t *testing.T) {
	for path, want := range map[string]string{
		"golang.org/x/time/rate": "/v1/files/golang.org/x/time/rate",
		"x%2Fy":                  "/v1/files/x%252Fy",
		"50%":                    "/v1/files/50%25",
		"../admin":               "/v1/files/%2E%2E/admin",
		"a/./b":                  "/v1/files/a/%2E/b",
	} {
		t.Run(path, func(t *testing.T) {
			if got := sentPath(t, func(c *Client) error { return c.GetFile(t.Context(), path) }); got != want {
				t.Errorf("sent to %s, want %s", got, want)
			}

			svc := &fileService{}
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

			if err := c.GetFile(t.Context(), path); err != nil {
				t.Fatal(err)
			}

			if svc.got != path {
				t.Errorf("the service got %q, want %q", svc.got, path)
			}
		})
	}
}
