package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
)

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
}
