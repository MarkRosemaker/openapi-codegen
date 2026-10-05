// This file is written by hand, not by the generator.
//
// It pins tagged structs beyond Notion's plainest shape: an alternative with
// members of its own beside the one named after its tag, the tag's enum type,
// and an allOf whose union holds a tagged struct among its alternatives.

package tagged

import (
	"encoding/json/v2"
	"strings"
	"testing"
)

func TestTagged_OwnMembers(t *testing.T) {
	var b Block
	if err := json.Unmarshal([]byte(`{"id":"b","type":"relation","relation":["p"],"has_more":true}`), &b, jsonOpts); err != nil {
		t.Fatal(err)
	}

	if b.Type != BlockTypeRelation || len(b.Relation) != 1 || b.HasMore == nil || !*b.HasMore {
		t.Errorf("got %+v, want the relation with has_more", b)
	}

	// has_more is optional where it belongs, and refused elsewhere
	if err := json.Unmarshal([]byte(`{"id":"b","type":"relation","relation":[]}`), &b, jsonOpts); err != nil {
		t.Errorf("relation without has_more: %v", err)
	}

	for in, want := range map[string]string{
		`{"id":"b","type":"paragraph","paragraph":"hi","has_more":true}`: `type "paragraph" does not allow member "has_more"`,
		`{"id":"b","type":"relation","has_more":true}`:                   `missing object member name "relation"`,
	} {
		if err := json.Unmarshal([]byte(in), &b, jsonOpts); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v, want %s", in, err, want)
		}
	}

	// left empty, the tag is the value whose member is set, has_more beside it or not
	out, err := json.Marshal(&Block{ID: "b", Relation: []string{"p"}, HasMore: new(false)}, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != `{"id":"b","type":"relation","relation":["p"],"has_more":false}` {
		t.Errorf("got %s", out)
	}
}

func TestTagged_Enum(t *testing.T) {
	var b Block
	if err := json.Unmarshal([]byte(`{"id":"b","type":"child_database","child_database":"Tasks"}`), &b, jsonOpts); err != nil {
		t.Fatal(err)
	}

	if b.Type != BlockTypeChildDatabase || !b.Type.Valid() || BlockType("table").Valid() {
		t.Errorf("got %q", b.Type)
	}
}

func TestTagged_InAllOfUnion(t *testing.T) {
	// the tagged struct two unions deep is chosen by the members present, as an object would be
	var v Value
	if err := json.Unmarshal([]byte(`{"id":"v","type":"relation","relation":["p"],"has_more":true}`), &v, jsonOpts); err != nil {
		t.Fatal(err)
	}

	s := v.ValueAllOf1.SimpleOrArray
	if v.ID != "v" || s == nil || s.Type != SimpleOrArrayTypeRelation || s.HasMore == nil || v.ValueAllOf1.Rollup != nil {
		t.Fatalf("got %+v, want the relation", v)
	}

	if err := json.Unmarshal([]byte(`{"id":"v","type":"rollup","function":"sum"}`), &v, jsonOpts); err != nil {
		t.Fatal(err)
	}

	if r := v.ValueAllOf1.Rollup; r == nil || r.Function != "sum" || v.ValueAllOf1.SimpleOrArray != nil {
		t.Fatalf("got %+v, want the rollup", v)
	}

	out, err := json.Marshal(&v, jsonOpts)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != `{"id":"v","type":"rollup","function":"sum"}` {
		t.Errorf("got %s", out)
	}

	// and is checked as it is anywhere
	if err := json.Unmarshal([]byte(`{"id":"v","type":"number","number":1,"text":"x"}`), &v, jsonOpts); err == nil ||
		!strings.Contains(err.Error(), `type "number" does not allow member "text"`) {
		t.Errorf("got %v, want text refused", err)
	}
}
