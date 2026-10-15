package ir_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

// unixDoc is a document whose component schemas are those given, one referred to by an operation.
func unixDoc(schemas map[string]*openapi.Schema) *openapi.Document {
	doc := &openapi.Document{
		OpenAPI: "3.1.0",
		Info:    &openapi.Info{Title: "T", Version: "1"},
		Servers: openapi.Servers{{URL: "https://example.com"}},
	}

	doc.Components.Schemas = openapi.Schemas{}
	for name, s := range schemas {
		doc.Components.Schemas.Set(name, s)
	}

	op := &openapi.Operation{OperationID: "listEvents"}
	op.Responses = openapi.OperationResponses{}
	op.Responses.Set("200", makeResponse("OK", "", nil))

	doc.Paths = openapi.Paths{}
	doc.Paths.Set("/events", &openapi.PathItem{Get: op})

	return doc
}

// object is an object schema requiring all of props, given as name and schema in turn.
func object(props ...any) *openapi.Schema {
	s := &openapi.Schema{Type: openapi.TypeObject, Properties: openapi.Schemas{}}

	for i := 0; i < len(props); i += 2 {
		name := props[i].(string)
		s.Properties.Set(name, props[i+1].(*openapi.Schema))
		s.Required = append(s.Required, name)
	}

	return s
}

func intDate() *openapi.Schema {
	return &openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatDate}
}

func strDate() *openapi.Schema {
	return &openapi.Schema{Type: openapi.TypeString, Format: openapi.FormatDate}
}

func TestUnix_AllIntegers(t *testing.T) {
	t.Parallel()

	// every date an integer: one marshaler writes them all, and no struct lists its own
	irDoc, err := ir.FromDocument(unixDoc(map[string]*openapi.Schema{
		"Event": object("raw", intDate(), "at", &openapi.Schema{Type: openapi.TypeInteger, Format: openapi.FormatDateTime}),
	}), "pkg", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if !irDoc.HasUnixDateFields || !irDoc.HasUnixTimeFields || irDoc.HasUnixFields() {
		t.Errorf("got package-wide date %t, time %t, per field %t; want true, false",
			irDoc.HasUnixDateFields, irDoc.HasUnixTimeFields, irDoc.HasUnixFields())
	}

	if got := irDoc.Schemas[0].Fields[0].Type; got != "civil.Date" {
		t.Errorf("an integer date is a %s, want civil.Date", got)
	}
}

func TestUnix_AmongStrings(t *testing.T) {
	t.Parallel()

	// dates of both forms: a struct lists its integer ones (golden covers those of embedded parts)
	irDoc, err := ir.FromDocument(unixDoc(map[string]*openapi.Schema{
		"Meta": object("expires", intDate(), "day", strDate()),
	}), "pkg", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if irDoc.HasUnixDateFields {
		t.Error("dates of both forms: got a marshaler for all of them")
	}

	want := map[string][]ir.UnixField{
		"Meta": {{Selector: "Expires", JSONTag: `json:"expires"`, Type: "jsonutil.DateIntUnix"}},
	}

	for _, s := range irDoc.Schemas {
		if !reflect.DeepEqual(s.Unix, want[s.Name]) {
			t.Errorf("%s: got %+v, want %+v", s.Name, s.Unix, want[s.Name])
		}
	}
}

func TestUnix_AmongStrings_Slice(t *testing.T) {
	t.Parallel()

	// a slice of integer dates among string ones has no field of its own to convert: it fails, saying why
	_, err := ir.FromDocument(unixDoc(map[string]*openapi.Schema{
		"Event": object("raws", &openapi.Schema{Type: openapi.TypeArray, Items: intDate()}, "day", strDate()),
	}), "pkg", "", false, false)
	if err == nil || !strings.Contains(err.Error(), "only supported as a value of its own") {
		t.Errorf("got %v, want an error saying why", err)
	}
}
