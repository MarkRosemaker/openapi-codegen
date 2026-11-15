// This file is written by hand, not by the generator.
//
// It pages through the lists the generator found to be paged by cursor: one by a query parameter and has_more, one
// by a member of the body and a cursor that is empty after the last page.

package golden

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"uuid"
)

// serve is a client of a server that answers with handle.
func serve(t *testing.T, handle http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(handle)
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

// notes is what the entries hold, in order.
func notes(t *testing.T, entries func(yield func(Entry, error) bool)) []string {
	t.Helper()

	var got []string

	for e, err := range entries {
		if err != nil {
			t.Fatal(err)
		}

		got = append(got, e.Note.Note)
	}

	return got
}

func TestListEntriesAll(t *testing.T) {
	next := uuid.New()

	var cursors []string

	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("start_cursor")
		cursors = append(cursors, cursor)

		w.Header().Set("Content-Type", "application/json")

		if cursor == "" {
			_, _ = w.Write([]byte(`{"results":[{"note":"a"},{"note":"b"}],"next_cursor":"` + next.String() + `","has_more":true}`))
		} else {
			_, _ = w.Write([]byte(`{"results":[{"note":"c"}],"next_cursor":null,"has_more":false}`))
		}
	})

	params := &ListEntriesParams{PageSize: 2}

	if got := notes(t, c.ListEntriesAll(t.Context(), params)); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("got %v, want a, b, c", got)
	}

	if !slices.Equal(cursors, []string{"", next.String()}) {
		t.Errorf("sent cursors %q, want none, then %s", cursors, next)
	}

	if params.StartCursor != uuid.Nil() {
		t.Error("paging changed the caller's params")
	}
}

func TestQueryEntriesAll(t *testing.T) {
	var sent []QueryEntries

	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		var body QueryEntries
		if err := json.UnmarshalRead(r.Body, &body); err != nil {
			t.Error(err)
		}

		sent = append(sent, body)

		w.Header().Set("Content-Type", "application/json")

		if body.Cursor == "" {
			_, _ = w.Write([]byte(`{"items":[{"note":"a"}],"next_page_token":"t2"}`))
		} else {
			_, _ = w.Write([]byte(`{"items":[{"note":"b"}],"next_page_token":""}`))
		}
	})

	if got := notes(t, c.QueryEntriesAll(t.Context(), QueryEntries{Kind: "todo"})); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("got %v, want a, b", got)
	}

	if want := []QueryEntries{{Kind: "todo"}, {Kind: "todo", Cursor: "t2"}}; !slices.Equal(sent, want) {
		t.Errorf("sent %v, want %v", sent, want)
	}
}
