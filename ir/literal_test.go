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
	}}

	for goType, want := range map[string]string{
		"Body":  "Body{User: User{ByID: new(ByID)}}",
		"User":  "User{ByID: new(ByID)}",
		"Plain": "Plain{}",
		"[]int": "[]int{}",
	} {
		if got := doc.MinimalLiteral(goType); got != want {
			t.Errorf("%s: got %s, want %s", goType, got, want)
		}
	}
}
