package ir_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_Tagged(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": {
    "Block": {"oneOf": [{"$ref": "#/components/schemas/Paragraph"}, {"$ref": "#/components/schemas/Divider"}, {"$ref": "#/components/schemas/Breadcrumb"}]},
    "Paragraph": {"type": "object", "required": ["id", "type", "paragraph"], "properties": {
      "id": {"type": "string"}, "type": {"type": "string", "const": "paragraph"}, "paragraph": {"$ref": "#/components/schemas/Text"}}},
    "Divider": {"type": "object", "required": ["id", "type"], "properties": {
      "id": {"type": "string", "description": "says another thing"}, "type": {"type": "string", "const": "divider"},
      "divider": {"type": "object", "properties": {}}}},
    "Breadcrumb": {"type": "object", "required": ["type"], "properties": {
      "id": {"type": "string"}, "type": {"type": "string", "const": "breadcrumb"}}},
    "Text": {"type": "object", "properties": {"text": {"type": "string"}}},

    "Parent": {"allOf": [{"$ref": "#/components/schemas/ParentCommon"}, {"$ref": "#/components/schemas/ParentUnion"}]},
    "ParentCommon": {"type": "object", "properties": {"note": {"type": "string"}}},
    "ParentUnion": {"oneOf": [
      {"type": "object", "required": ["type", "page_id"], "properties": {"type": {"const": "page_id", "type": "string"}, "page_id": {"type": "string"}}},
      {"type": "object", "required": ["type", "workspace"], "properties": {"type": {"const": "workspace", "type": "string"}, "workspace": {"type": "boolean"}}}
    ]},

    "Model": {"oneOf": [
      {"type": "object", "properties": {"mode": {"const": "auto", "type": "string"}}},
      {"type": "object", "properties": {"mode": {"const": "pinned", "type": "string"}, "id": {"type": "string"}}}
    ]},
    "Mixed": {"oneOf": [
      {"type": "object", "properties": {"type": {"const": "a", "type": "string"}, "id": {"type": "string"}}},
      {"type": "object", "properties": {"type": {"const": "b", "type": "string"}, "id": {"type": "integer"}}}
    ]}
  }}
}`))
	if err != nil {
		t.Fatal(err)
	}

	irDoc, err := ir.FromDocument(doc, "t", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	schemas := map[string]ir.Schema{}
	for _, s := range irDoc.Schemas {
		schemas[s.Name] = s
	}

	fields := func(s ir.Schema) []string {
		var out []string
		for _, f := range s.Fields {
			out = append(out, f.Name+" "+f.Type)
		}

		return out
	}

	// shared members, the tag, then one optional field per alternative's own member; a description does not count
	block := schemas["Block"]
	if block.Tagged == nil {
		t.Fatalf("Block is not tagged: %+v", block)
	}

	if got, want := fields(block), []string{"ID string", "Type BlockType", "Paragraph *Text", "Divider *struct{}"}; !slices.Equal(got, want) {
		t.Errorf("Block fields: got %q, want %q", got, want)
	}

	want := []ir.TaggedValue{
		{Value: "paragraph", Members: []ir.TaggedOwn{{Name: "paragraph", Required: true}}},
		{Value: "divider", Members: []ir.TaggedOwn{{Name: "divider"}}},
		{Value: "breadcrumb"},
	}
	if !reflect.DeepEqual(block.Tagged.Values, want) {
		t.Errorf("Block values: got %+v, want %+v", block.Tagged.Values, want)
	}

	// the tag is an enum of its values
	if got := block.Tagged.Enum; len(got) != 3 || got[0].GoName != "BlockTypeParagraph" || got[2].Value != "breadcrumb" {
		t.Errorf("Block enum: got %+v", got)
	}

	// its alternatives are needed by nothing else, so they get no type; Text is a member's type
	for _, name := range []string{"Paragraph", "Divider", "Breadcrumb"} {
		if _, ok := schemas[name]; ok {
			t.Errorf("%s still has a type", name)
		}
	}

	if _, ok := schemas["Text"]; !ok {
		t.Error("Text has no type")
	}

	// an allOf takes the union's fields in, and the union needs no type of its own
	parent := schemas["Parent"]
	if parent.Tagged == nil || parent.AllOfUnion != nil {
		t.Fatalf("Parent is not tagged: %+v", parent)
	}

	if got, want := fields(parent), []string{"Note string", "Type ParentType", "PageID string", "Workspace *bool"}; !slices.Equal(got, want) {
		t.Errorf("Parent fields: got %q, want %q", got, want)
	}

	if _, ok := schemas["ParentUnion"]; ok {
		t.Error("ParentUnion still has a type")
	}

	// a member not named after the tag's value, or a shared member that differs, keeps the union
	for _, name := range []string{"Model", "Mixed"} {
		if s := schemas[name]; s.Tagged != nil || s.Kind != ir.SchemaKindUnion {
			t.Errorf("%s: got %+v, want a union", name, s)
		}
	}
}
