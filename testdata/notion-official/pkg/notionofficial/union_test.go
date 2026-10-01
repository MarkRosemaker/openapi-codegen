// This file is written by hand, not by the generator.
//
// It pins how unions decode: by a discriminating member where one exists, which
// must come first, and, as part of an allOf, each part from the members it
// declares.

package notionofficial

import (
	"encoding/json/v2"
	"errors"
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
		`{"mode":"random"}`:                  `unknown value of "mode"`,
		`{"id":"gpt","mode":"pinned"}`:       `first member is "id", want "mode"`,
		`{"mode":"pinned","id":"gpt","x":1}`: `unknown object member name "x"`,
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
		`{"type":"page_id","page_id":"59833787-2cf9-4fdf-8782-e53db20768a5","extra":1}`: `unknown object member name "extra"`,
		`{"type":"page_id","workspace":true}`:                                           `unknown object member name "workspace"`,
		`{"page_id":"59833787-2cf9-4fdf-8782-e53db20768a5"}`:                            `first member is "page_id", want "type"`,
		`{"page_id":"59833787-2cf9-4fdf-8782-e53db20768a5","type":"page_id"}`:           `first member is "page_id", want "type"`,
		`{}`: `missing object member name "type"`,
	} {
		var p CreateDatabaseParent
		if err := json.Unmarshal([]byte(in), &p, jsonOpts); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %s", in, err, want)
		}
	}
}

func TestAllOf_NestedUnion(t *testing.T) {
	// the alternative is a leaf two unions deep, chosen by its type like any other
	const number = `{"type":"number","id":"abc","number":3}`

	var p pagePropertyValueWithIdResponse
	if err := json.Unmarshal([]byte(number), &p, jsonOpts); err != nil {
		t.Fatal(err)
	}

	value := p.PagePropertyValueWithIdResponseAllOf1.SimpleOrArrayPropertyValueResponse
	if p.ID != "abc" || value == nil || value.SimplePropertyValueResponse == nil ||
		value.SimplePropertyValueResponse.NumberFormulaPropertyValue == nil ||
		*value.SimplePropertyValueResponse.NumberFormulaPropertyValue.Number != 3 {
		t.Fatalf("got %+v, want id abc and the number 3 two unions deep", p)
	}

	out, err := json.Marshal(&p, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	// the discriminator comes first again, so the result decodes as well
	if string(out) != `{"type":"number","id":"abc","number":3}` {
		t.Errorf("got %s", out)
	}
}

func TestUnion_EncodesDiscriminatorFirst(t *testing.T) {
	// the alternative set decides the discriminator, so encoding writes it in first
	out, err := json.Marshal(&AgentModel{AgentModelOneOf2: &AgentModelOneOf2{ID: new("gpt")}}, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != `{"mode":"pinned","id":"gpt"}` {
		t.Errorf("got %s", out)
	}

	_, err = json.Marshal(&AgentModel{AgentModelOneOf2: &AgentModelOneOf2{Mode: "auto", ID: new("gpt")}}, jsonOpts)
	if err == nil || !strings.Contains(err.Error(), `member "mode" is "auto", want "pinned"`) {
		t.Errorf("got %v, want the wrong mode refused", err)
	}
}

func TestUnion_StandardErrors(t *testing.T) {
	// an unknown member reads, and matches, as encoding/json's own
	var m AgentModel
	err := json.Unmarshal([]byte(`{"mode":"pinned","id":"gpt","x":1}`), &m, jsonOpts)

	var se *json.SemanticError
	if !errors.As(err, &se) || !errors.Is(err, json.ErrUnknownName) || se.JSONPointer != "/x" {
		t.Errorf("got %v, want a semantic error for an unknown name at /x", err)
	}
}
