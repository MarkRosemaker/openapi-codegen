package ir

import (
	"testing"

	"github.com/MarkRosemaker/openapi"
)

func TestEscapePathSegment(t *testing.T) {
	for in, want := range map[string]string{
		"abc-_.~09": "abc-_.~09",
		"C^C:":      "C%5EC%3A",
		"a b+c":     "a%20b%2Bc",
		"a/b;c@d":   "a%2Fb%3Bc%40d",
		"50%":       "50%25",
		".":         "%2E",
		"..":        "%2E%2E",
		"...":       "...",
	} {
		if got := escapePathSegment(in); got != want {
			t.Errorf("escapePathSegment(%q) = %q, want %q", in, got, want)
		}
	}

	if got, want := escapePathSegments("a b/../c:d"), "a%20b/%2E%2E/c%3Ad"; got != want {
		t.Errorf("escapePathSegments: got %q, want %q", got, want)
	}
}

func TestEscapesPathSegment(t *testing.T) {
	doc := func(params ...Param) Document {
		return Document{Operations: []Operation{{PathParams: params}}}
	}

	for _, tc := range []struct {
		name string
		doc  Document
		want bool
	}{
		{"none", doc(), false},
		{"a UUID", doc(Param{Type: "uuid.UUID"}), false},
		{"a fixed value", doc(Param{Type: "string", Value: `"a:b"`}), false},
		{"a string", doc(Param{Type: "string"}), true},
		{"slashes kept", doc(Param{Type: "string", AllowSlashes: true}), true},
	} {
		if got := tc.doc.EscapesPathSegment(); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}

	if got, want := (Param{Type: "string", Value: `"a:b"`}).PathExpr(), `"a%3Ab"`; got != want {
		t.Errorf("a fixed value: got %s, want %s", got, want)
	}
}

func TestZeroUnencodable(t *testing.T) {
	union := &openapi.Schema{AnyOf: openapi.SchemaList{
		{Type: openapi.TypeString, Format: openapi.FormatDate},
		{Type: openapi.TypeString, Format: openapi.FormatDateTime},
	}}
	ref := func(name string, s *openapi.Schema) *openapi.Schema {
		return &openapi.Schema{Ref: &openapi.SchemaRef{Identifier: "#/components/schemas/" + name, Value: s}}
	}
	object := func(required []string, props openapi.Schemas) *openapi.Schema {
		return &openapi.Schema{Type: openapi.TypeObject, Properties: props, Required: required}
	}

	date := object([]string{"start"}, openapi.Schemas{"start": ref("DateOrDateTime", union), "end": union})
	cyclic := object([]string{"self"}, openapi.Schemas{})
	cyclic.Properties["self"] = ref("Cyclic", cyclic)

	for _, tc := range []struct {
		name string
		s    *openapi.Schema
		want bool
	}{
		{"requires a union", date, true},
		{"by reference", ref("Date", date), true},
		{"through an object it requires", object([]string{"date"}, openapi.Schemas{"date": date}), true},
		{"through allOf", &openapi.Schema{AllOf: openapi.SchemaList{date}}, true},
		{"the union only optional", object(nil, openapi.Schemas{"start": union}), false},
		{"the union nullable", object([]string{"start"}, openapi.Schemas{
			"start": {OneOf: openapi.SchemaList{union, {Type: openapi.TypeNull}}},
		}), false},
		{"requires a string", object([]string{"name"}, openapi.Schemas{"name": {Type: openapi.TypeString}}), false},
		{"a string", &openapi.Schema{Type: openapi.TypeString}, false},
		{"its own member", cyclic, false},
		{"nil", nil, false},
	} {
		if got := zeroUnencodable(tc.s); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}

	tp, err := nullableGoType(ref("Date", date))
	if err != nil {
		t.Fatal(err)
	}

	if !tp.IsPointer {
		t.Error("a nullable object whose zero value cannot be encoded is no pointer")
	}
}

func TestNullableGoType(t *testing.T) {
	one := 1.0
	union := &openapi.Schema{OneOf: openapi.SchemaList{
		{Type: openapi.TypeString},
		{Type: openapi.TypeObject, Required: []string{"id"}, Properties: openapi.Schemas{"id": {Type: openapi.TypeString}}},
	}}

	for _, tc := range []struct {
		name    string
		s       *openapi.Schema
		pointer bool
	}{
		// null reads as the zero value: no pointer
		{"a string", &openapi.Schema{Type: openapi.TypeString}, false},
		{"a time", &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDateTime}, false},
		{"a date", &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDate}, false},
		{"an ID from 1", &openapi.Schema{Type: openapi.TypeInteger, Min: &one}, false},
		{"an object requiring nothing", &openapi.Schema{Type: openapi.TypeObject, Properties: openapi.Schemas{
			"name": {Type: openapi.TypeString},
		}}, false},
		{"a number", &openapi.Schema{Type: openapi.TypeInteger}, false},
		{"a boolean", &openapi.Schema{Type: openapi.TypeBoolean}, false},
		{"a union, which reads null as no alternative", union, false},
		// it cannot: a pointer
		{"a tuple, which refuses null", &openapi.Schema{Type: openapi.TypeArray, PrefixItems: openapi.SchemaList{{Type: openapi.TypeString}}}, true},
	} {
		tp, err := nullableGoType(tc.s)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}

		if tp.IsPointer != tc.pointer {
			t.Errorf("%s: %s is a pointer: %v, want %v", tc.name, tp.Name, tp.IsPointer, tc.pointer)
		}
	}
}
