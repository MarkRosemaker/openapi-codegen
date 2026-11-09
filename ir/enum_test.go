package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

// An enum is checked with each member, and with a value of its type that none is, where there is one: for a number,
// the smallest whole number from 0 up that no member holds, which fits every number type.
func TestDocument_Enums(t *testing.T) {
	t.Parallel()

	enum := func(name, tp string, values ...string) ir.Schema {
		s := ir.Schema{Name: name, Kind: ir.SchemaKindEnum, Type: tp}
		for _, v := range values {
			s.EnumValues = append(s.EnumValues, ir.EnumValue{GoName: name + v, Value: v})
		}

		return s
	}

	doc := ir.Document{Schemas: []ir.Schema{
		enum("Color", "string", "red"),
		enum("Taken", "string", "not a member"),
		enum("Level", "int", "3", "1"),
		enum("Counted", "int", "0", "1", "2"),
		enum("Negative", "int", "-5", "0"),
		enum("Max32", "int32", "2147483647"),                              // one past it would not compile
		enum("Big", "int64", "9007199254740993", "9007199254740992", "0"), // a float would take the two for one
		enum("Byte", "uint8", "0"),
		enum("Ratio", "float64", "0", "0.5", "1.0"),
		enum("Flag", "bool", "true"),
		{Name: "None", Kind: ir.SchemaKindEnum, Type: "string"},
		{Name: "Page", Kind: ir.SchemaKindStruct},
	}}

	want := map[string]string{
		"Color":    `"not a member"`,
		"Taken":    `"not a member!"`, // a member holds the first choice
		"Level":    "0",
		"Counted":  "3",
		"Negative": "1",
		"Max32":    "0",
		"Big":      "1",
		"Byte":     "1",
		"Ratio":    "2", // 1.0 is the member 1
		"Flag":     "",  // both values may be members
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
