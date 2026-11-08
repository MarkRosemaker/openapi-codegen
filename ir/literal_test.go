package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestDocument_MinimalLiteral(t *testing.T) {
	t.Parallel()

	doc := ir.Document{Schemas: []ir.Schema{
		{Name: "Body", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
			{Name: "User", Type: "User", Required: true},
			{Name: "Note", Type: "string", Required: true},
			{Name: "Other", Type: "*User"},
		}},
		{Name: "User", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{
			{FieldName: "ByID", Type: "ByID"}, {FieldName: "String", Type: "string"},
		}},
		{Name: "ByID", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "ID", Type: "string", Required: true}}},
		{Name: "Plain", Kind: ir.SchemaKindStruct},
		{Name: "Trial", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "End", Type: "civil.Date", Required: true}}},
		{Name: "Wrapper", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{{FieldName: "Body", Type: "Body"}}},
	}}

	for goType, want := range map[string]string{
		"Body":  "Body{User: User{ByID: new(ByID)}}",
		"User":  "User{ByID: new(ByID)}",
		"Plain": "Plain{}",
		// a variant that sets something is a pointer to its literal
		"Wrapper": "Wrapper{Body: &Body{User: User{ByID: new(ByID)}}}",
		"[]int":   "[]int{}",
		// the zero date marshals as 0000-00-00, which does not parse back
		"Trial": "Trial{End: civil.Date{Year: 1, Month: 1, Day: 1}}",
	} {
		if got := doc.MinimalLiteral(goType); got != want {
			t.Errorf("%s: got %s, want %s", goType, got, want)
		}
	}
}
