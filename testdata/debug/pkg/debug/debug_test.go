// This file is written by hand, not by the generator.
//
// It pins how a client in debug mode decodes what the specification does not
// know yet: it records the interaction, then decodes again without rejecting
// unknown members, and fails only if that fails too.

package debug

import (
	"encoding/json/jsontext"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi-enrich/cassette"
	"github.com/go-api-libs/api"
)

func serve(t *testing.T, body string) *url.URL {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

// recorded returns the interactions the client wrote to api/interactions.json.
func recorded(t *testing.T) cassette.Interactions {
	t.Helper()

	ias, err := cassette.InteractionsReadFile("api/interactions.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		t.Fatal(err)
	}

	return ias
}

// sameJSON reports whether got and want are the same JSON value, however written.
func sameJSON(t *testing.T, got []byte, want string) bool {
	t.Helper()

	g, w := jsontext.Value(slices.Clone(got)), jsontext.Value(want)
	if err := g.Canonicalize(); err != nil {
		t.Fatal(err)
	}

	if err := w.Canonicalize(); err != nil {
		t.Fatal(err)
	}

	return string(g) == string(w)
}

func TestDebug_Lenient(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		call       func(*Client) (any, error)
		check      func(any) bool
	}{
		{
			name: "struct",
			body: `{"id":"a","new":1}`,
			call: func(c *Client) (any, error) { return c.GetItem(t.Context()) },
			check: func(v any) bool {
				return v.(*Item).ID == "a"
			},
		},
		{
			// unknown members inside the member the tag names, and beside it
			name: "tagged union",
			body: `{"type":"paragraph","paragraph":{"text":"hi","color":"red"},"new":true}`,
			call: func(c *Client) (any, error) { return c.GetBlock(t.Context()) },
			check: func(v any) bool {
				b := v.(*Block)
				return b.Type == "paragraph" && b.Paragraph != nil && b.Paragraph.Text == "hi"
			},
		},
		{
			name: "union",
			body: `{"meow":true,"new":1}`,
			call: func(c *Client) (any, error) { return c.GetPet(t.Context()) },
			check: func(v any) bool {
				p := v.(*Pet)
				return p.Cat != nil && p.Cat.Meow && p.Dog == nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			strict, err := NewClient(WithBaseURL(serve(t, tc.body)), WithBearer("token"))
			if err != nil {
				t.Fatal(err)
			}

			if _, err := tc.call(strict); err == nil {
				t.Fatal("without debug: decoded, want the unknown member refused")
			} else if _, ok := errors.AsType[*api.DecodingError](err); !ok {
				t.Fatalf("without debug: got %T, want *api.DecodingError", err)
			}

			if ias := recorded(t); len(ias) != 0 {
				t.Fatalf("without debug, %d interactions recorded", len(ias))
			}

			c, err := NewClient(WithBaseURL(serve(t, tc.body)), WithBearer("token"), WithDebug)
			if err != nil {
				t.Fatal(err)
			}

			v, err := tc.call(c)
			if err != nil {
				t.Fatalf("with debug: %v", err)
			}

			if !tc.check(v) {
				t.Errorf("got %+v", v)
			}

			if ias := recorded(t); len(ias) != 1 || !sameJSON(t, ias[0].Response.Body, tc.body) {
				t.Errorf("recorded %+v, want the one response", ias)
			}
		})
	}
}

func TestDebug_StillFails(t *testing.T) {
	t.Chdir(t.TempDir())

	// a member of the wrong type is no member the specification does not know, so it fails even leniently
	c, err := NewClient(WithBaseURL(serve(t, `{"id":1}`)), WithBearer("token"), WithDebug)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.GetItem(t.Context()); err == nil {
		t.Fatal("decoded, want an error")
	} else if _, ok := errors.AsType[*api.DecodingError](err); !ok {
		t.Errorf("got %T, want *api.DecodingError", err)
	}

	if ias := recorded(t); len(ias) != 1 {
		t.Errorf("recorded %d interactions, want 1", len(ias))
	}
}

func TestDebug_FileOnceMasked(t *testing.T) {
	path := t.TempDir() + "/recorded.json"

	c, err := NewClient(WithBaseURL(serve(t, `{"id":"a","password":"hunter2"}`)), WithBearer("token"), WithDebugFile(path))
	if err != nil {
		t.Fatal(err)
	}

	// polled, a call that fails the same way is added once, to the file asked for, and with its credentials masked
	for range 2 {
		if item, err := c.GetItem(t.Context()); err != nil || item.ID != "a" {
			t.Fatalf("got %+v, %v", item, err)
		}
	}

	ias, err := cassette.InteractionsReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(ias) != 1 {
		t.Fatalf("recorded %d interactions, want 1", len(ias))
	}

	if auth := ias[0].Request.Headers.Get("Authorization"); auth != "Bearer *****" || !sameJSON(t, ias[0].Response.Body, `{"id":"a","password":"*******"}`) {
		t.Errorf("recorded %s and %s, want both masked", auth, ias[0].Response.Body)
	}
}
