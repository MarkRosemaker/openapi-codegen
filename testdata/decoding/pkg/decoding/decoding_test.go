// This file is written by hand, not by the generator.
//
// It pins decoding of shapes the Notion client met in real responses: a tag
// left out where it is optional, null for an "X or null" union, an anyOf of a
// full and a partial form, a caller's own type of a result, and a union with
// a tuple among its alternatives.

package decoding

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const page = `{"object":"page","id":"p","cover":null,"icon":null,"position":["a","b"]}`

func client(t *testing.T, body string) *Client {
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

	c, err := NewClient(WithBaseURL(u))
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestDecoding_OptionalTag(t *testing.T) {
	// Notion's own filters leave type out; it is the value whose member is set
	var f PropertyFilter
	if err := json.Unmarshal([]byte(`{"property":"x","select":{"does_not_equal":"Done"}}`), &f, jsonOpts); err != nil {
		t.Fatal(err)
	}

	if f.Type != PropertyFilterTypeSelect || f.Select == nil || f.Select.DoesNotEqual != "Done" {
		t.Errorf("got %+v, want the select filter", f)
	}

	// without a member to tell by, it is still missing
	if err := json.Unmarshal([]byte(`{"property":"x"}`), &f, jsonOpts); err == nil {
		t.Error("decoded a filter without type or member")
	}
}

func TestDecoding_NullableUnion(t *testing.T) {
	p, err := client(t, page).GetPage(t.Context(), "p")
	if err != nil {
		t.Fatal(err)
	}

	if p.Cover != nil || p.Icon != nil {
		t.Errorf("got %+v, want no cover and no icon", p)
	}

	p, err = client(t, `{"object":"page","id":"p","cover":{"type":"external","external":{"url":"u"}},"icon":{"emoji":"x"},"position":["a",1]}`).
		GetPage(t.Context(), "p")
	if err != nil {
		t.Fatal(err)
	}

	if p.Cover == nil || p.Cover.External == nil || p.Icon == nil || p.Icon.Emoji == nil || p.Position.Pair == nil {
		t.Errorf("got %+v, want the cover, the icon and the pair", p)
	}

	out, err := json.Marshal(p, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != `{"object":"page","id":"p","cover":{"type":"external","external":{"url":"u"}},"icon":{"emoji":"x"},"position":["a",1]}` {
		t.Errorf("got %s", out)
	}
}

func TestDecoding_FullOrPartial(t *testing.T) {
	// the full page matches; the partial form, strict, does not
	results, err := client(t, `[`+page+`,{"object":"page","id":"q"}]`).Search(t.Context(), PropertyFilter{Property: "x", Select: &Condition{}})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 || results[0].Page == nil || results[0].PartialPage != nil || results[1].PartialPage == nil {
		t.Errorf("got %+v", results)
	}
}

func TestDecoding_OwnResultType(t *testing.T) {
	// a type of the caller's own takes what it declares of the response and leaves the rest
	p, err := client(t, page).GetPageWithResult[struct {
		ID string `json:"id"`
	}](t.Context(), "p")
	if err != nil {
		t.Fatal(err)
	}

	if p.ID != "p" {
		t.Errorf("got %+v", p)
	}

	// the operation's own type is decoded as strictly as ever
	if _, err := client(t, `{"object":"page","id":"p","cover":null,"icon":null,"position":[],"extra":1}`).GetPage(t.Context(), "p"); err == nil {
		t.Error("decoded an unknown member into Page")
	}
}
