package ir_test

import (
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

// A union's variant is round-tripped where a value of it can be built and no other variant would decode it too.
func TestDocument_UnionRoundTrips(t *testing.T) {
	t.Parallel()

	doc := ir.Document{Schemas: []ir.Schema{
		{Name: "Circle", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
			{Name: "Kind", JSONName: "kind", Type: "string", Required: true},
			{Name: "Radius", JSONName: "radius", Type: "int", Required: true},
		}},
		{Name: "Square", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
			{Name: "Kind", JSONName: "kind", Type: "string", Required: true},
			{Name: "Side", JSONName: "side", Type: "int", Required: true},
		}},
		{Name: "Partial", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "ID", JSONName: "id", Type: "string", Required: true}}},
		{Name: "Full", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
			{Name: "ID", JSONName: "id", Type: "string", Required: true},
			{Name: "Name", JSONName: "name", Type: "string", Required: true},
		}},
		{Name: "Named", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
			{Name: "ID", JSONName: "id", Type: "string", Required: true},
			{Name: "Name", JSONName: "name", Type: "string"},
		}},
		{Name: "Dated", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "On", JSONName: "on", Type: "civil.Date", Required: true}}},
		{Name: "Images", Kind: ir.SchemaKindAlias, Type: "[]Partial"},
		{Name: "Code", Kind: ir.SchemaKindAlias, Type: "string"},
		// told apart by a discriminator: every variant, with the discriminator set
		{Name: "Shape", Kind: ir.SchemaKindUnion, Discriminator: "kind", UnionVariants: []ir.UnionVariant{
			{FieldName: "Circle", Type: "Circle", Value: "circle", Object: true, Members: []string{"kind", "radius"}, Required: []string{"kind", "radius"}},
			{FieldName: "Square", Type: "Square", Value: "square", Object: true, Members: []string{"kind", "side"}, Required: []string{"kind", "side"}},
		}},
		// Full requires a name Partial does not declare, so a minimal value of either is one of it alone
		{Name: "PageOrPartial", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "Full", Type: "Full", Object: true, Members: []string{"id", "name"}, Required: []string{"id", "name"}},
			{FieldName: "Partial", Type: "Partial", Object: true, Members: []string{"id"}, Required: []string{"id"}},
		}},
		// Named's name is optional: a minimal Named is {"id": …}, as a minimal Partial is, so neither is told apart
		{Name: "NamedOrPartial", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "Named", Type: "Named", Object: true, Members: []string{"id", "name"}, Required: []string{"id"}},
			{FieldName: "Partial", Type: "Partial", Object: true, Members: []string{"id"}, Required: []string{"id"}},
		}},
		// two strings are the same kind of value; an array and a string are not
		{Name: "Codes", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "String", Type: "string", Zero: `""`},
			{FieldName: "Code", Type: "Code", Zero: `""`},
		}},
		{Name: "OneOrMany", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "String", Type: "string", Zero: `""`},
			{FieldName: "Images", Type: "Images", Zero: "nil"},
		}},
		// a date's zero value is no date
		{Name: "WhenOrCode", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "Dated", Type: "Dated", Object: true, Members: []string{"on"}, Required: []string{"on"}},
			{FieldName: "Int", Type: "int"},
		}},
	}}

	got := map[string][]string{}
	for _, u := range doc.UnionRoundTrips() {
		for _, v := range u.Variants {
			got[u.Union] = append(got[u.Union], v.FieldName)
		}
	}

	want := map[string][]string{
		"Shape":         {"Circle", "Square"},
		"PageOrPartial": {"Full", "Partial"},
		"OneOrMany":     {"String", "Images"},
		"WhenOrCode":    {"Int"},
	}

	for union, fields := range want {
		if !slices.Equal(got[union], fields) {
			t.Errorf("%s: got %v, want %v", union, got[union], fields)
		}
	}

	for union, why := range map[string]string{"Codes": "both are strings", "NamedOrPartial": "their minimal values are alike"} {
		if len(got[union]) > 0 {
			t.Errorf("%s: got %v, want none, as %s", union, got[union], why)
		}
	}

	for _, u := range doc.UnionRoundTrips() {
		if u.Union != "Shape" {
			continue
		}

		if v := u.Variants[0]; !slices.Contains(v.Setup, `v.Kind = "circle"`) {
			t.Errorf("Circle: setup %v sets no discriminator", v.Setup)
		}
	}
}

// A minimal value sets a union's first variant, held by value where the variant is a named slice: not a pointer to it.
func TestDocument_MinimalLiteralNamedSliceVariant(t *testing.T) {
	t.Parallel()

	doc := ir.Document{Schemas: []ir.Schema{
		{Name: "Tools", Kind: ir.SchemaKindAlias, Type: "[]string"},
		{Name: "Enabled", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{{FieldName: "Tools", Type: "Tools", Zero: "nil"}}},
	}}

	if got, want := doc.MinimalLiteral("Enabled"), "Enabled{Tools: Tools{}}"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
