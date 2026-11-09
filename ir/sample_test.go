package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

var sampleDoc = ir.Document{Schemas: []ir.Schema{
	{Name: "Level", Kind: ir.SchemaKindEnum, EnumValues: []ir.EnumValue{
		{GoName: "LevelOne", Value: "1"}, {GoName: "LevelThree", Value: "3"},
	}},
	{Name: "Empty", Kind: ir.SchemaKindEnum},
	{Name: "Page", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "ID", Type: "string", Required: true}}},
	{Name: "Pet", Kind: ir.SchemaKindUnion, UnionVariants: []ir.UnionVariant{{FieldName: "Cat", Type: "Cat"}}},
	{Name: "Owner", Kind: ir.SchemaKindStruct, Tagged: &ir.Tagged{}},
	{Name: "Home", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
		{Name: "Pet", Type: "Pet", Required: true},
	}},
	{Name: "Yard", Kind: ir.SchemaKindStruct, Fields: []ir.Field{
		{Name: "Pet", Type: "*Pet"}, {Name: "Owner", Type: "Owner"},
	}},
	{Name: "Porch", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "Pet", Type: "*Pet", Required: true}}},
	{Name: "Deed", Kind: ir.SchemaKindStruct, Fields: []ir.Field{{Name: "Owner", Type: "Owner", Required: true}}},
	{Name: "Message", Kind: ir.SchemaKindAlias, Type: "string"},
}}

// Every parameter that can be found in the request again is sampled, each with a value of its own.
func TestDocument_ParamSamples(t *testing.T) {
	t.Parallel()

	op := ir.Operation{
		QueryParams: ir.Params{
			{FieldName: "Name", JSONName: "name", Type: "string"},
			{FieldName: "Other", JSONName: "other", Type: "string"},
			{FieldName: "Limit", JSONName: "limit", Type: "int"},
			{FieldName: "Offset", JSONName: "offset", Type: "int64"},
			{FieldName: "Ratio", JSONName: "ratio", Type: "float64"},
			{FieldName: "Verbose", JSONName: "verbose", Type: "bool"},
			{FieldName: "Level", JSONName: "level", Type: "Level", IsEnum: true},
			{FieldName: "Code", JSONName: "code", Type: "Code", BaseType: "string"},
			{FieldName: "IDs", JSONName: "ids", Type: "[]int", Item: &ir.Param{Type: "int"}},
			// none of these can be sampled
			{FieldName: "Since", JSONName: "since", Type: "time.Time"},
			{FieldName: "None", JSONName: "none", Type: "Empty", IsEnum: true},
			{FieldName: "Filter", JSONName: "filter", Type: "Filter", Props: ir.Params{{Type: "string"}}},
			{FieldName: "Tags", JSONName: "tags", Type: "map[string]string", MapValue: &ir.Param{Type: "string"}},
		},
		HeaderParams: ir.Params{
			{FieldName: "Version", JSONName: "X-Version", Type: "string"},
			{FieldName: "Accept", JSONName: "Accept", Type: "[]string", Item: &ir.Param{Type: "string"}},
		},
	}

	type sample struct{ field, in, literal, wire string }

	want := []sample{
		{"Name", "query", `"name"`, "name"},
		{"Other", "query", `"other"`, "other"},
		{"Limit", "query", "4", "4"},
		{"Offset", "query", "5", "5"},
		{"Ratio", "query", "6.5", "6.5"},
		{"Verbose", "query", "true", "true"},
		// the last member, since the first is the likeliest default
		{"Level", "query", "LevelThree", "3"},
		{"Code", "query", `Code("code")`, "code"},
		{"IDs", "query", "[]int{10}", "10"},
		{"Version", "header", `"X-Version"`, "X-Version"},
	}

	got := sampleDoc.ParamSamples(op)
	if len(got) != len(want) {
		t.Fatalf("got %d samples, want %d: %+v", len(got), len(want), got)
	}

	for i, w := range want {
		g := got[i]
		if (sample{g.FieldName, g.In, g.Literal, g.Wire}) != w {
			t.Errorf("sample %d: got %s %s = %s, sent as %q; want %s %s = %s, sent as %q",
				i, g.In, g.FieldName, g.Literal, g.Wire, w.in, w.field, w.literal, w.wire)
		}
	}

	if !(ir.Document{Operations: []ir.Operation{op}}).HasParamSamples() {
		t.Error("an operation with samples: want HasParamSamples")
	}

	if (ir.Document{Operations: []ir.Operation{{}}}).HasParamSamples() {
		t.Error("an operation without parameters: want no HasParamSamples")
	}
}
