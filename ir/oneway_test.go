package ir_test

import (
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_OneWay(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": {
    "User": {"type": "object", "required": ["id", "name", "password"], "properties": {
      "id": {"type": "string", "readOnly": true}, "name": {"type": "string"}, "password": {"type": "string", "writeOnly": true}}},
    "Meta": {"type": "object", "required": ["created"], "properties": {"created": {"type": "string", "readOnly": true}}},
    "Record": {"allOf": [{"$ref": "#/components/schemas/Meta"}, {"type": "object", "properties": {"record": {"type": "string"}}}]},
    "Note": {"allOf": [{"$ref": "#/components/schemas/Meta"}, {"type": "object", "required": ["note"], "properties": {"note": {"type": "string"}}}]},
    "Entry": {"oneOf": [{"$ref": "#/components/schemas/Record"}, {"$ref": "#/components/schemas/Note"}]}
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

	// neither direction requires what only one carries
	for _, f := range schemas["User"].Fields {
		if want := f.Name == "Name"; f.Required != want {
			t.Errorf("User.%s: required %t, want %t", f.Name, f.Required, want)
		}
	}

	// the fields to leave out, through an embedded part too
	if got := schemas["User"].ReadOnly; !slices.Equal(got, []string{"ID"}) {
		t.Errorf("User read-only: got %q", got)
	}

	if got := schemas["User"].WriteOnly; !slices.Equal(got, []string{"Password"}) {
		t.Errorf("User write-only: got %q", got)
	}

	if got := schemas["Record"].ReadOnly; !slices.Equal(got, []string{"Meta.Created"}) {
		t.Errorf("Record read-only: got %q", got)
	}

	// a union's alternative is chosen without them
	var required []string
	for _, v := range schemas["Entry"].UnionVariants {
		required = append(required, v.Required...)
	}

	if !slices.Equal(required, []string{"note"}) {
		t.Errorf("Entry requires %q, want note", required)
	}

	if !irDoc.HasReadOnly() || !irDoc.HasWriteOnly() {
		t.Error("the document has no one-way fields")
	}
}
