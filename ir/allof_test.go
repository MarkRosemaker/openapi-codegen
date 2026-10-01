package ir_test

import (
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func loadSchemas(t *testing.T, components string) *ir.Document {
	t.Helper()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": ` + components + `}
}`))
	if err != nil {
		t.Fatal(err)
	}

	irDoc, err := ir.FromDocument(doc, "t", "", false)
	if err != nil {
		t.Fatal(err)
	}

	return irDoc
}

func schemaNamed(t *testing.T, doc *ir.Document, name string) *ir.Schema {
	t.Helper()

	i := slices.IndexFunc(doc.Schemas, func(s ir.Schema) bool { return s.Name == name })
	if i < 0 {
		return nil
	}

	return &doc.Schemas[i]
}

func TestFromDocument_UnionDiscriminator(t *testing.T) {
	t.Parallel()

	doc := loadSchemas(t, `{
		"ByConst": {"oneOf": [{"$ref": "#/components/schemas/Cat"}, {"$ref": "#/components/schemas/Dog"}]},
		"ByMapping": {
			"oneOf": [{"$ref": "#/components/schemas/Cat"}, {"$ref": "#/components/schemas/Dog"}],
			"discriminator": {"propertyName": "kind", "mapping": {"meow": "Cat"}}
		},
		"Undecided": {"oneOf": [{"$ref": "#/components/schemas/Cat"}, {"$ref": "#/components/schemas/Kitten"}]},
		"Cat": {"type": "object", "properties": {"type": {"type": "string", "const": "cat"}, "kind": {"type": "string"}}},
		"Dog": {"type": "object", "properties": {"type": {"type": "string", "const": "dog"}, "kind": {"type": "string"}}},
		"Kitten": {"type": "object", "properties": {"type": {"type": "string", "const": "cat"}}}
	}`)

	for name, want := range map[string]struct {
		discriminator string
		values        []string
	}{
		"ByConst":   {"type", []string{"cat", "dog"}},
		"ByMapping": {"kind", []string{"meow", "Dog"}},
		"Undecided": {"", []string{"", ""}},
	} {
		s := schemaNamed(t, doc, name)

		var values []string
		for _, v := range s.UnionVariants {
			values = append(values, v.Value)
		}

		if s.Discriminator != want.discriminator || !slices.Equal(values, want.values) {
			t.Errorf("%s: got %q by %v, want %q by %v", name, s.Discriminator, values, want.discriminator, want.values)
		}
	}
}

func TestFromDocument_AllOfParts(t *testing.T) {
	t.Parallel()

	doc := loadSchemas(t, `{
		"Pet": {"allOf": [
			{"$ref": "#/components/schemas/PetAllOf0"},
			{"$ref": "#/components/schemas/Named"},
			{"$ref": "#/components/schemas/PetAllOf2"}
		]},
		"Toy": {"allOf": [{"$ref": "#/components/schemas/Named"}]},
		"PetAllOf0": {"type": "object", "properties": {"id": {"type": "string"}}, "required": ["id"]},
		"Named": {"type": "object", "properties": {"name": {"type": "string"}}},
		"PetAllOf2": {"oneOf": [{"$ref": "#/components/schemas/Cat"}, {"$ref": "#/components/schemas/Dog"}]},
		"Cat": {"type": "object", "properties": {"type": {"type": "string", "const": "cat"}, "meow": {"type": "boolean"}}},
		"Dog": {"type": "object", "properties": {"type": {"type": "string", "const": "dog"}}},
		"Twice": {"allOf": [{"$ref": "#/components/schemas/PetAllOf2"}, {"$ref": "#/components/schemas/PetAllOf2"}]}
	}`)

	// a part only Pet uses is folded into its fields and needs no type; a shared one stays embedded
	if schemaNamed(t, doc, "PetAllOf0") != nil {
		t.Error("PetAllOf0 is still a type of its own")
	}

	pet := schemaNamed(t, doc, "Pet")

	var fields []string
	for _, f := range pet.Fields {
		fields = append(fields, f.Name+"|"+f.Type+"|"+f.JSONTag)
	}

	want := []string{`ID|string|json:"id,omitzero"`, "|Named|", `PetAllOf2|PetAllOf2|json:"-"`}
	if !slices.Equal(fields, want) {
		t.Errorf("Pet fields: got %q, want %q", fields, want)
	}

	if !pet.Fields[1].Embedded || !slices.Equal(pet.Members, []string{"id", "name"}) {
		t.Errorf("Pet: got embedded %t and members %v, want Named embedded and members id, name", pet.Fields[1].Embedded, pet.Members)
	}

	// the union is a field of its own, its alternatives told apart by type
	u := pet.AllOfUnion
	if u == nil || u.FieldName != "PetAllOf2" || u.Discriminator != "type" || len(u.Variants) != 2 ||
		!slices.Equal(u.Variants[0].Members, []string{"meow", "type"}) {
		t.Errorf("Pet: got union %+v, want PetAllOf2 by type with Cat's members", u)
	}

	if got := schemaNamed(t, doc, "Twice").Unimplemented; got == "" {
		t.Error("Twice: an allOf of two unions is not marked unimplemented")
	}
}
