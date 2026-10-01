// This file is written by hand, not by the generator.
//
// It pins how unions decode: by a discriminating member where one exists, which
// must come first, and, as part of an allOf, each part from the members it
// declares.

package notionofficial

import (
	"encoding/json/v2"
	"strings"
	"testing"
)

func TestUnion_Discriminator(t *testing.T) {
	var m AgentModel
	if err := json.Unmarshal([]byte(`{"mode":"pinned","id":"gpt"}`), &m, jsonOpts); err != nil {
		t.Fatal(err)
	}

	if m.AgentModelOneOf2 == nil || m.AgentModelOneOf != nil {
		t.Errorf("got %+v, want only the pinned alternative", m)
	}

	for in, want := range map[string]string{
		`{"mode":"random"}`:                  `unknown mode "random"`,
		`{"id":"gpt","mode":"pinned"}`:       `first member is "id", want "mode"`,
		`{"mode":"pinned","id":"gpt","x":1}`: `unknown member "x"`,
	} {
		if err := json.Unmarshal([]byte(in), &m, jsonOpts); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %s", in, err, want)
		}
	}
}

func TestAllOf_Union(t *testing.T) {
	const page = `{"type":"page_id","page_id":"59833787-2cf9-4fdf-8782-e53db20768a5"}`

	var p CreateDatabaseParent
	if err := json.Unmarshal([]byte(page), &p, jsonOpts); err != nil {
		t.Fatal(err)
	}

	// the field outside the union and the chosen alternative both get the members they declare
	if p.Type != "page_id" || p.CreateDatabaseParentAllOf2.CreateACommentAllOfOneOfParentOneOf == nil {
		t.Errorf("got %+v, want type page_id and the page alternative", p)
	}

	out, err := json.Marshal(&p, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != page {
		t.Errorf("got %s, want %s", out, page)
	}

	for in, want := range map[string]string{
		`{"type":"page_id","page_id":"59833787-2cf9-4fdf-8782-e53db20768a5","extra":1}`: `unknown member "extra"`,
		`{"type":"page_id","workspace":true}`:                                           `unknown member "workspace"`,
		`{"page_id":"59833787-2cf9-4fdf-8782-e53db20768a5"}`:                            `first member is "page_id", want "type"`,
		`{"page_id":"59833787-2cf9-4fdf-8782-e53db20768a5","type":"page_id"}`:           `first member is "page_id", want "type"`,
		`{}`: `missing member "type"`,
	} {
		var p CreateDatabaseParent
		if err := json.Unmarshal([]byte(in), &p, jsonOpts); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %s", in, err, want)
		}
	}
}
