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

// Only a failure decoded into a type of its own, under a status code of its own, is an error response.
func TestOperation_ErrorResponses(t *testing.T) {
	t.Parallel()

	op := ir.Operation{Responses: ir.Responses{
		{StatusCode: "200", ContentType: "application/json", GoType: &ir.GoType{Name: "Page"}, IsSuccess: true},
		{StatusCode: "404", ContentType: "application/json", GoType: &ir.GoType{Name: "Error"}},
		{StatusCode: "4XX", ContentType: "application/json", GoType: &ir.GoType{Name: "Error"}},
		{StatusCode: "default", ContentType: "application/json", GoType: &ir.GoType{Name: "Error"}},
		{StatusCode: "409", ContentType: "application/json", GoType: &ir.GoType{Name: "Error", IsSlice: true}},
		{StatusCode: "410", ContentType: "text/plain", GoType: &ir.GoType{Name: "[]byte"}, IsRawBytes: true},
		{StatusCode: "500"},
	}}

	got := op.ErrorResponses()
	if len(got) != 1 || got[0].StatusCode != "404" {
		t.Fatalf("got %+v, want the 404 alone", got)
	}

	if !(ir.Document{Operations: []ir.Operation{op}}).HasErrorResponses() {
		t.Error("want HasErrorResponses")
	}
}

// A success decoded into a struct of the package's own is strict outside debug mode and lenient in it, unless the
// struct holds a union its minimal literal cannot be decoded back from.
func TestDocument_StrictAndLenientSuccess(t *testing.T) {
	t.Parallel()

	success := func(name string) ir.Operation {
		return ir.Operation{
			SuccessReturn: &ir.GoType{Name: name},
			Responses: ir.Responses{
				{StatusCode: "404", ContentType: "application/json", GoType: &ir.GoType{Name: "Page"}},
				{StatusCode: "200", ContentType: "application/json", GoType: &ir.GoType{Name: name}, IsSuccess: true},
			},
		}
	}

	for name, want := range map[string]bool{
		"Page": true,
		// a union, or a struct that requires one, tagged or not
		"Pet":  false,
		"Home": false,
		"Deed": false,
		// a union behind a pointer or in an optional field is left out of the minimal literal
		"Yard":    true,
		"Porch":   true,
		"Message": false,
		"Missing": false,
	} {
		op := success(name)

		if got := sampleDoc.StrictSuccess(op) != nil; got != want {
			t.Errorf("%s: strict %t, want %t", name, got, want)
		}

		debug := sampleDoc
		debug.Debug = true

		if debug.StrictSuccess(op) != nil {
			t.Errorf("%s: strict in debug mode", name)
		}

		if got := debug.LenientSuccess(op) != nil; got != want {
			t.Errorf("%s: lenient in debug mode %t, want %t", name, got, want)
		}

		if sampleDoc.LenientSuccess(op) != nil {
			t.Errorf("%s: lenient outside debug mode", name)
		}
	}

	for name, op := range map[string]ir.Operation{
		"no success value": {Responses: ir.Responses{{StatusCode: "204", IsSuccess: true}}},
		"empty success":    func() ir.Operation { op := success("Page"); op.EmptySuccess = true; return op }(),
		"no success":       {SuccessReturn: &ir.GoType{Name: "Page"}},
		"slice":            {SuccessReturn: &ir.GoType{Name: "Page"}, Responses: ir.Responses{{StatusCode: "200", ContentType: "application/json", GoType: &ir.GoType{Name: "Page", IsSlice: true}, IsSuccess: true}}},
	} {
		if sampleDoc.StrictSuccess(op) != nil {
			t.Errorf("%s: want no strict success", name)
		}
	}

	if !(ir.Document{Schemas: sampleDoc.Schemas, Operations: []ir.Operation{success("Page")}}).HasStructSuccess() {
		t.Error("want HasStructSuccess")
	}
}

func TestDocument_ValueLiteral(t *testing.T) {
	t.Parallel()

	for goType, want := range map[string]string{
		"Page": `Page{}`,
		// a named string, which a composite literal cannot build
		"Message": "*new(Message)",
		"Level":   "*new(Level)",
		"Missing": "*new(Missing)",
	} {
		if got := sampleDoc.ValueLiteral(goType); got != want {
			t.Errorf("%s: got %s, want %s", goType, got, want)
		}
	}
}

// An enum is checked with each member, and with a value of its type that none is, where there is one.
func TestDocument_Enums(t *testing.T) {
	t.Parallel()

	doc := ir.Document{Schemas: []ir.Schema{
		{Name: "Color", Kind: ir.SchemaKindEnum, Type: "string", EnumValues: []ir.EnumValue{{GoName: "ColorRed", Value: "red"}}},
		{Name: "Taken", Kind: ir.SchemaKindEnum, Type: "string", EnumValues: []ir.EnumValue{{GoName: "TakenNot", Value: "not a member"}}},
		{Name: "Level", Kind: ir.SchemaKindEnum, Type: "int", EnumValues: []ir.EnumValue{{GoName: "LevelThree", Value: "3"}, {GoName: "LevelOne", Value: "1"}}},
		{Name: "Ratio", Kind: ir.SchemaKindEnum, Type: "float64", EnumValues: []ir.EnumValue{{GoName: "RatioHalf", Value: "0.5"}, {GoName: "RatioMore", Value: "2.5"}}},
		{Name: "Flag", Kind: ir.SchemaKindEnum, Type: "bool", EnumValues: []ir.EnumValue{{GoName: "FlagTrue", Value: "true"}}},
		{Name: "None", Kind: ir.SchemaKindEnum, Type: "string"},
		{Name: "Page", Kind: ir.SchemaKindStruct},
	}}

	want := map[string]string{
		"Color": `"not a member"`,
		"Taken": `"not a member!"`, // a member holds the first choice
		"Level": "4",
		"Ratio": "3", // a whole number, which fits every number type
		"Flag":  "",  // both values may be members
	}

	enums := doc.Enums()
	if len(enums) != len(want) {
		t.Fatalf("got %d enums, want %d: an enum without members and a struct are left out", len(enums), len(want))
	}

	for _, e := range enums {
		if got := e.NotAMember(); got != want[e.Name] {
			t.Errorf("%s: not a member is %s, want %s", e.Name, got, want[e.Name])
		}
	}
}

// Debug mode's failure paths are tested where it decodes: a JSON success, an error struct, and a failure that declares
// no body is reported by its status.
func TestDocument_DebugAndBodilessResponses(t *testing.T) {
	t.Parallel()

	op := ir.Operation{
		SuccessReturn: &ir.GoType{Name: "Page"},
		Responses: ir.Responses{
			{StatusCode: "200", ContentType: "application/json", GoType: &ir.GoType{Name: "Page"}, IsSuccess: true},
			{StatusCode: "404", ContentType: "application/json", GoType: &ir.GoType{Name: "Page"}},
			{StatusCode: "409"},
			{StatusCode: "4XX"},
			{StatusCode: "204", IsSuccess: true},
		},
	}

	debug := sampleDoc
	debug.Debug = true
	debug.Operations = []ir.Operation{op}

	if r := debug.DebugDecoded(op); r == nil || r.StatusCode != "200" {
		t.Errorf("in debug mode: got %+v, want the 200", r)
	}

	if sampleDoc.DebugDecoded(op) != nil {
		t.Error("outside debug mode: want none")
	}

	for name, o := range map[string]ir.Operation{
		"empty success":  {EmptySuccess: true, SuccessReturn: &ir.GoType{Name: "Page"}, Responses: op.Responses},
		"no value":       {Responses: op.Responses},
		"bytes, no JSON": {SuccessReturn: &ir.GoType{Name: "[]byte"}, Responses: ir.Responses{{StatusCode: "200", ContentType: "text/plain", IsRawBytes: true, IsSuccess: true}}},
	} {
		if debug.DebugDecoded(o) != nil {
			t.Errorf("%s: want none", name)
		}
	}

	for goType, want := range map[string]bool{"Page": true, "Home": false, "Message": false} {
		r := ir.Response{StatusCode: "404", ContentType: "application/json", GoType: &ir.GoType{Name: goType}}
		if got := debug.LenientError(r); got != want {
			t.Errorf("%s in debug mode: lenient %t, want %t", goType, got, want)
		}

		if sampleDoc.LenientError(r) {
			t.Errorf("%s outside debug mode: want not lenient", goType)
		}
	}

	if got := op.BodilessFailures(); len(got) != 1 || got[0].StatusCode != "409" {
		t.Errorf("got %+v, want the 409 alone", got)
	}

	if !debug.NeedsBodyServer() || !debug.NeedsUnknownMember() {
		t.Error("want both helpers")
	}

	if (ir.Document{}).NeedsBodyServer() || (ir.Document{}).NeedsUnknownMember() {
		t.Error("no operations: want neither helper")
	}
}
