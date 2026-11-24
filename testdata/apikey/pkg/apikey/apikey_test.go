// This file is written by hand, not by the generator.
//
// It pins how a client sends an API key a security scheme declares: read from the environment or given with
// WithAPIKey, sent in the header the scheme names on every operation that takes it, and only there, and never written
// to what debug mode records.

package apikey

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const key = "k3y-of-the-quotes"

// fakeQuotes answers every request, the quote with a member the specification does not know where unknown, which
// debug mode records, and keeps what each path was sent.
func fakeQuotes(t *testing.T, unknown bool) (*url.URL, func(path string) *http.Request) {
	t.Helper()

	var (
		mu   sync.Mutex
		sent = map[string]*http.Request{}
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		sent[r.URL.Path] = r
		mu.Unlock()

		if r.URL.Path != "/quote" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		body := `{"c": 12.3}`
		if unknown {
			body = `{"c": 12.3, "t": 1700000000}`
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	return u, func(path string) *http.Request {
		mu.Lock()
		defer mu.Unlock()

		return sent[path]
	}
}

func TestNewClient_NoKey(t *testing.T) {
	t.Setenv("QUOTES_API_KEY", "")

	if _, err := NewClient(); err == nil || !strings.Contains(err.Error(), "QUOTES_API_KEY") {
		t.Fatalf("got %v, want an error naming QUOTES_API_KEY", err)
	}
}

func TestAPIKey_SentWhereTheSchemeSays(t *testing.T) {
	for name, opts := range map[string][]ClientOption{
		"from the environment": nil,
		"given":                {WithAPIKey(key)},
	} {
		t.Run(name, func(t *testing.T) {
			if opts == nil {
				t.Setenv("QUOTES_API_KEY", key)
			} else {
				t.Setenv("QUOTES_API_KEY", "")
			}

			u, sent := fakeQuotes(t, false)

			c, err := NewClient(append(opts, WithBaseURL(u))...)
			if err != nil {
				t.Fatal(err)
			}

			if q, err := c.GetQuote(t.Context()); err != nil {
				t.Fatal(err)
			} else if q.C != 12.3 {
				t.Errorf("price %v, want 12.3", q.C)
			}

			if err := c.GetStatus(t.Context()); err != nil {
				t.Fatal(err)
			}

			if err := c.Health(t.Context()); err != nil {
				t.Fatal(err)
			}

			for path, want := range map[string]string{"/quote": key, "/status": key, "/health": ""} {
				r := sent(path)
				if got := r.Header.Get("X-Quotes-Token"); got != want {
					t.Errorf("%s: X-Quotes-Token %q, want %q", path, got, want)
				}

				if strings.Contains(r.URL.RawQuery, key) {
					t.Errorf("%s: the key is in the query too: %s", path, r.URL.RawQuery)
				}
			}
		})
	}
}

func TestAPIKey_NeverRecorded(t *testing.T) {
	t.Setenv("QUOTES_API_KEY", "")

	for name, opt := range map[string]func(path string) ClientOption{
		// the quote holds a member the specification does not know, which debug mode records
		"debug":     WithDebugFile,
		"recording": func(path string) ClientOption { return func(c *Client) { WithRecording(c); WithDebugFile(path)(c) } },
	} {
		t.Run(name, func(t *testing.T) {
			u, _ := fakeQuotes(t, true)
			path := filepath.Join(t.TempDir(), "interactions.json")

			c, err := NewClient(WithAPIKey(key), WithBaseURL(u), opt(path))
			if err != nil {
				t.Fatal(err)
			}

			if _, err := c.GetQuote(t.Context()); err != nil {
				t.Fatal(err)
			}

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("nothing recorded: %v", err)
			}

			if strings.Contains(string(b), key) {
				t.Errorf("the key is recorded:\n%s", b)
			}

			if !strings.Contains(string(b), "X-Quotes-Token") {
				t.Errorf("the header is not recorded at all, masked:\n%s", b)
			}
		})
	}
}
