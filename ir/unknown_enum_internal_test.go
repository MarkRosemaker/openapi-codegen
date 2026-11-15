package ir

import "testing"

func TestEnumsCheck(t *testing.T) {
	t.Parallel()

	e := newEnums([]Schema{
		{Name: "Status", Kind: SchemaKindEnum, Type: "string"},
		{Name: "Level", Kind: SchemaKindEnum, Type: "int"},
		{Name: "Item", Kind: SchemaKindStruct, Fields: []Field{{Name: "Status", Type: "Status"}}},
		{Name: "Items", Kind: SchemaKindAlias, Type: "[]Item"},
		{Name: "Same", Kind: SchemaKindAlias, Type: "Items", IsTypeAlias: true},
		{Name: "Node", Kind: SchemaKindStruct, Fields: []Field{{Name: "Next", Type: "*Node"}, {Name: "Name", Type: "string"}}},
	})

	for _, tc := range []struct {
		typ      string
		optional bool
		want     string
	}{
		{"string", false, ""},
		{"Node", false, ""}, // refers to itself, but holds no enum
		{"Status", false, "!v.Valid()"},
		{"Level", true, "v != 0 && !v.Valid()"},
		{"*Item", false, "v != nil && v.unknownEnum()"},
		{"*[]Status", false, "v != nil && slices.ContainsFunc((*v), func(x Status) bool { return !x.Valid() })"},
		{"[][]Status", false, "slices.ContainsFunc(v, func(x []Status) bool { return " +
			"slices.ContainsFunc(x, func(x2 Status) bool { return !x2.Valid() }) })"},
		{"map[Status]string", false, "func() bool {\nfor k := range v {\nif !k.Valid() {\nreturn true\n}\n}\n\nreturn false\n}()"},
		{"map[string][]Item", false, "func() bool {\nfor _, x := range v {\nif slices.ContainsFunc(x, func(x2 Item) bool { return x2.unknownEnum() }) {\nreturn true\n}\n}\n\nreturn false\n}()"},
		{"Same", false, "v.unknownEnum()"},
	} {
		if got := e.check("v", tc.typ, tc.optional, 0); got != tc.want {
			t.Errorf("check(%s) =\n%s\nwant\n%s", tc.typ, got, tc.want)
		}
	}
}

func TestMarkUnknownEnums(t *testing.T) {
	t.Parallel()

	schemas := []Schema{
		{Name: "Status", Kind: SchemaKindEnum, Type: "string"},
		{Name: "Item", Kind: SchemaKindStruct, Fields: []Field{{Name: "Status", Type: "Status", Required: true}}},
		{Name: "Page", Kind: SchemaKindStruct, Fields: []Field{{Name: "Results", Type: "[]Item"}}},
		{Name: "Problem", Kind: SchemaKindStruct, Fields: []Field{{Name: "Code", Type: "Status"}}},
		{Name: "Filter", Kind: SchemaKindStruct, Fields: []Field{{Name: "Status", Type: "Status"}}},
	}

	markUnknownEnums(schemas, []Operation{{
		SuccessReturn: &GoType{Name: "Page", IsPointer: true},
		Responses:     Responses{{GoType: &GoType{Name: "Page"}, IsSuccess: true}, {GoType: &GoType{Name: "Problem"}}},
		RequestBody:   &ReqBody{TypeName: "Filter"},
	}})

	for _, s := range schemas {
		want := map[string]string{
			"Item":    `v.Status != "" && !v.Status.Valid()`,
			"Page":    "slices.ContainsFunc(v.Results, func(x Item) bool { return x.unknownEnum() })",
			"Problem": `v.Code != "" && !v.Code.Valid()`,
		}[s.Name] // a request's filter is never checked, so it has none

		if s.UnknownEnum != want {
			t.Errorf("%s: UnknownEnum = %q, want %q", s.Name, s.UnknownEnum, want)
		}
	}
}
