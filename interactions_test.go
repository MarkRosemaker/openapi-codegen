package codegen

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
)

func TestMatchInteractions_SkipsBadURL(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{Name: "GetFoo", Method: "GET", PathTemplate: "/foo"},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "://invalid"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err == nil {
		t.Fatal("expected error")
	} else if got, want := err.Error(), `parse "://invalid": missing protocol scheme`; got != want {
		t.Fatalf("got error=`%s`, want=`%s`", got, want)
	}
}

func TestMatchInteractions_ErrorsOnMethodMismatch(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{Name: "PostFoo", Method: "POST", PathTemplate: "/foo"},
		},
	}
	err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 200},
		},
	})
	if err == nil {
		t.Fatal("expected error for unmatched interaction")
	}
}

func TestMatchInteractions_ErrorsOnNoMatch(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{Name: "GetBar", Method: "GET", PathTemplate: "/bar"},
		},
	}
	err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 200},
		},
	})
	if err == nil {
		t.Fatal("expected error for unmatched interaction")
	}
}

func TestMatchInteractions_WithPathParam(t *testing.T) {
	doc := &ir.Document{
		BaseURL: ir.URLParts{Path: "/v1"},
		Operations: []ir.Operation{
			{
				Name:         "GetPet",
				Method:       "GET",
				PathTemplate: "/{id}",
				PathParams: []ir.Param{
					{GoName: "id", JSONName: "id", Type: "string"},
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://api.example.com/v1/abc-123"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	if len(doc.InteractionCalls[0].PathArgs) != 1 {
		t.Fatalf("expected 1 path arg, got %d", len(doc.InteractionCalls[0].PathArgs))
	}
	if doc.InteractionCalls[0].PathArgs[0] != `"abc-123"` {
		t.Errorf("unexpected path arg: %s", doc.InteractionCalls[0].PathArgs[0])
	}
}

func TestMatchInteractions_WithQueryParam(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{
				Name:            "ListItems",
				Method:          "GET",
				PathTemplate:    "/items",
				ParamStructName: "ListItemsParams",
				QueryParams: []ir.Param{
					{
						GoName:    "limit",
						FieldName: "Limit",
						JSONName:  "limit",
						Type:      "int",
					},
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/items?limit=10"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	if len(doc.InteractionCalls[0].QueryArgs) != 1 {
		t.Fatalf("expected 1 query arg, got %d", len(doc.InteractionCalls[0].QueryArgs))
	}
	if doc.InteractionCalls[0].QueryArgs[0].Literal != "10" {
		t.Errorf("unexpected query arg literal: %s", doc.InteractionCalls[0].QueryArgs[0].Literal)
	}
}

func TestMatchInteractions_EmptyBasePath(t *testing.T) {
	doc := &ir.Document{
		// BaseURL.Path is empty: base == root
		Operations: []ir.Operation{
			{Name: "GetFoo", Method: "GET", PathTemplate: "/foo"},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(doc.InteractionCalls) != 1 {
		t.Errorf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
}

func TestMatchPathTemplate(t *testing.T) {
	tests := []struct {
		tmpl  string
		path  string
		match bool
		key   string
		val   string
	}{
		{"/foo", "/foo", true, "", ""},
		{"/foo", "/bar", false, "", ""},
		{"/{id}", "/abc", true, "id", "abc"},
		{"/CIK{cik}.json", "/CIK0000320193.json", true, "cik", "0000320193"},
		{"/{id}/children", "/abc/children", true, "id", "abc"},
		{"/{id}/children", "/abc/grandchildren", false, "", ""},
		{"/a/b", "/a", false, "", ""},
		// trailing wildcard param consuming multiple slash-separated segments
		{"/package/{path}", "/package/github.com/google/go-cmp/cmp", true, "path", "github.com/google/go-cmp/cmp"},
		{"/package/{path}", "/package/single", true, "path", "single"},
	}

	for _, tc := range tests {
		vals, ok, _ := matchPathTemplate(tc.tmpl, tc.path)
		if ok != tc.match {
			t.Errorf("matchPathTemplate(%q, %q): got match=%v, want %v", tc.tmpl, tc.path, ok, tc.match)
			continue
		}
		if tc.key != "" && vals[tc.key] != tc.val {
			t.Errorf("matchPathTemplate(%q, %q): param[%q]=%q, want %q",
				tc.tmpl, tc.path, tc.key, vals[tc.key], tc.val)
		}
	}
}

func TestMatchPathTemplate_ExactFlag(t *testing.T) {
	// Segment-for-segment matches are exact; wildcard captures are not.
	if _, ok, exact := matchPathTemplate("/tasks/{id}", "/tasks/abc"); !ok || !exact {
		t.Errorf("segment-count match: ok=%v exact=%v, want true true", ok, exact)
	}
	if _, ok, exact := matchPathTemplate("/tasks/{id}", "/tasks/abc/score/up"); !ok || exact {
		t.Errorf("wildcard match: ok=%v exact=%v, want true false", ok, exact)
	}
	if _, ok, exact := matchPathTemplate("/tasks/{id}/score/up", "/tasks/abc/score/up"); !ok || !exact {
		t.Errorf("longer exact match: ok=%v exact=%v, want true true", ok, exact)
	}
}

func TestMatchInteractions_PrefersExactOverWildcard(t *testing.T) {
	// Both operations match /tasks/abc/score/up: the shorter template wildcards
	// the trailing segments, but the longer template matches exactly. The
	// longer one must win regardless of declaration order.
	scoreUp := ir.Operation{
		Name:         "ListApiv3TaskScoreUp",
		Method:       "GET",
		PathTemplate: "/tasks/{taskId}/score/up",
		PathParams:   []ir.Param{{GoName: "taskID", JSONName: "taskId", Type: "string"}},
	}
	getTask := ir.Operation{
		Name:         "GetApiv3TaskByTaskID",
		Method:       "GET",
		PathTemplate: "/tasks/{taskId}",
		PathParams:   []ir.Param{{GoName: "taskID", JSONName: "taskId", Type: "string"}},
	}

	for _, order := range [][]ir.Operation{{getTask, scoreUp}, {scoreUp, getTask}} {
		doc := &ir.Document{Operations: order}
		if err := matchInteractions(doc, cassette.Interactions{
			{
				Request:  cassette.Request{Method: "GET", URL: "https://habitica.com/tasks/abc/score/up"},
				Response: cassette.Response{StatusCode: 200},
			},
		}); err != nil {
			t.Fatal(err)
		}
		if len(doc.InteractionCalls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
		}
		if got := doc.InteractionCalls[0].Op.Name; got != "ListApiv3TaskScoreUp" {
			t.Errorf("order %v: matched %s, want ListApiv3TaskScoreUp", []string{order[0].Name, order[1].Name}, got)
		}
		if got := doc.InteractionCalls[0].PathArgs[0]; got != `"abc"` {
			t.Errorf("path arg = %s, want %q", got, `"abc"`)
		}
	}
}

func TestExtractSegmentParam_MidSegmentMismatch(t *testing.T) {
	out := map[string]string{}
	// Prefix mismatch
	if extractSegmentParam("CIK{id}.json", "WRONG0001.json", out) {
		t.Error("expected mismatch on prefix")
	}
	// Suffix mismatch
	if extractSegmentParam("CIK{id}.json", "CIK0001.xml", out) {
		t.Error("expected mismatch on suffix")
	}
	// Both match
	if !extractSegmentParam("CIK{id}.json", "CIK0001.json", out) {
		t.Error("expected match")
	}
}

func TestGoLiteralForType(t *testing.T) {
	tests := []struct {
		goType string
		value  string
		want   string
	}{
		{"string", "hello", `"hello"`},
		{"int", "42", "42"},
		{"int32", "10", "10"},
		{"int64", "100", "100"},
		{"uint", "5", "5"},
		{"uint32", "3", "3"},
		{"uint64", "7", "7"},
		{
			"uuid.UUID", "96245c8f-1784-44a4-82ad-1941127c3ec3",
			`uuid.MustParse("96245c8f-1784-44a4-82ad-1941127c3ec3")`,
		},
		{"float64", "1.5", `"1.5"`},
	}
	for _, tc := range tests {
		got := goLiteralForType(tc.goType, tc.value)
		if got != tc.want {
			t.Errorf("goLiteralForType(%q, %q) = %q, want %q", tc.goType, tc.value, got, tc.want)
		}
	}
}

func TestExtractSegmentParam_NoParam(t *testing.T) {
	out := map[string]string{}
	if !extractSegmentParam("foo", "foo", out) {
		t.Error("expected match for static segment")
	}
	if extractSegmentParam("foo", "bar", out) {
		t.Error("expected mismatch for static segment")
	}
}

func TestMatchInteractions_MidSegmentParam(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{
				Name:         "GetCompany",
				Method:       "GET",
				PathTemplate: "/api/xbrl/companyfacts/CIK{cik}.json",
				PathParams: []ir.Param{
					{GoName: "cik", JSONName: "cik", Type: "string"},
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request: cassette.Request{
				Method: "GET",
				URL:    "https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json",
			},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	if len(doc.InteractionCalls[0].PathArgs) != 1 {
		t.Fatalf("expected 1 path arg, got %d", len(doc.InteractionCalls[0].PathArgs))
	}
	want := `"0000320193"`
	if doc.InteractionCalls[0].PathArgs[0] != want {
		t.Errorf("path arg = %q, want %q", doc.InteractionCalls[0].PathArgs[0], want)
	}
}

func TestMatchInteractions_BasePathStripping(t *testing.T) {
	doc := &ir.Document{
		BaseURL: ir.URLParts{Path: "/v1/pets"},
		Operations: []ir.Operation{
			{Name: "ListPets", Method: "GET", PathTemplate: "/"},
			{
				Name:         "GetPet",
				Method:       "GET",
				PathTemplate: "/{petId}",
				PathParams:   []ir.Param{{GoName: "petID", JSONName: "petId", Type: "string"}},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "https://api.petstoreapi.com/v1/pets"},
			Response: cassette.Response{StatusCode: 200},
		},
		{
			Request:  cassette.Request{Method: "GET", URL: "https://api.petstoreapi.com/v1/pets/abc-123"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if len(doc.InteractionCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(doc.InteractionCalls))
	}
	if doc.InteractionCalls[0].Op.Name != "ListPets" {
		t.Errorf("expected ListPets, got %s", doc.InteractionCalls[0].Op.Name)
	}
	if doc.InteractionCalls[1].Op.Name != "GetPet" {
		t.Errorf("expected GetPet, got %s", doc.InteractionCalls[1].Op.Name)
	}
	if !strings.Contains(doc.InteractionCalls[1].PathArgs[0], "abc-123") {
		t.Errorf("expected abc-123 in path arg, got %s", doc.InteractionCalls[1].PathArgs[0])
	}
}

func TestMatchInteractions_SuccessStatusCode(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{
				Name:         "GetFoo",
				Method:       "GET",
				PathTemplate: "/foo",
				Responses: ir.Responses{
					{StatusCode: "200", IsSuccess: true, GoType: &ir.GoType{Name: "Foo"}},
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	call := doc.InteractionCalls[0]
	if !call.IsSuccess {
		t.Error("IsSuccess = false, want true")
	}
	if call.ErrorType != "" {
		t.Errorf("ErrorType = %q, want empty", call.ErrorType)
	}
}

func TestMatchInteractions_ErrorStatusCodeWithSchema(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{
				Name:         "PostFoo",
				Method:       "POST",
				PathTemplate: "/foo",
				Responses: ir.Responses{
					{StatusCode: "200", IsSuccess: true, GoType: &ir.GoType{Name: "Foo"}},
					{StatusCode: "404", GoType: &ir.GoType{Name: "FooNotFoundResponse"}},
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "POST", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 404},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	call := doc.InteractionCalls[0]
	if call.IsSuccess {
		t.Error("IsSuccess = true, want false")
	}
	if call.ErrorType != "FooNotFoundResponse" {
		t.Errorf("ErrorType = %q, want FooNotFoundResponse", call.ErrorType)
	}
}

func TestMatchInteractions_ErrorStatusCodeWithoutSchema(t *testing.T) {
	doc := &ir.Document{
		Operations: []ir.Operation{
			{
				Name:         "PostFoo",
				Method:       "POST",
				PathTemplate: "/foo",
				Responses: ir.Responses{
					{StatusCode: "200", IsSuccess: true, GoType: &ir.GoType{Name: "Foo"}},
					{StatusCode: "403"}, // declared but no schema
				},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "POST", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 403},
		},
	}); err != nil {
		t.Fatal(err)
	}
	call := doc.InteractionCalls[0]
	if call.IsSuccess {
		t.Error("IsSuccess = true, want false")
	}
	if call.ErrorType != "" {
		t.Errorf("ErrorType = %q, want empty", call.ErrorType)
	}
}

func TestMatchInteractions_UndeclaredStatusCodeFallback(t *testing.T) {
	// No Responses declared at all: fall back to the HTTP convention.
	doc := &ir.Document{
		Operations: []ir.Operation{
			{Name: "GetFoo", Method: "GET", PathTemplate: "/foo"},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "http://example.com/foo"},
			Response: cassette.Response{StatusCode: 500},
		},
	}); err != nil {
		t.Fatal(err)
	}
	call := doc.InteractionCalls[0]
	if call.IsSuccess {
		t.Error("IsSuccess = true, want false for 500 fallback")
	}
	if call.ErrorType != "" {
		t.Errorf("ErrorType = %q, want empty", call.ErrorType)
	}
}

func TestFindResponse(t *testing.T) {
	op := &ir.Operation{
		Responses: ir.Responses{
			{StatusCode: "200", IsSuccess: true},
			{StatusCode: "404"},
		},
	}
	if r := findResponse(op, 200); r == nil || !r.IsSuccess {
		t.Errorf("findResponse(200) = %v, want success response", r)
	}
	if r := findResponse(op, 404); r == nil || r.IsSuccess {
		t.Errorf("findResponse(404) = %v, want non-success response", r)
	}
	if r := findResponse(op, 500); r != nil {
		t.Errorf("findResponse(500) = %v, want nil", r)
	}
}

func TestMatchInteractions_MultiSegmentPathParam(t *testing.T) {
	doc := &ir.Document{
		BaseURL: ir.URLParts{Path: "/v1beta"},
		Operations: []ir.Operation{
			{
				Name:         "GetPackage",
				Method:       "GET",
				PathTemplate: "/package/{path}",
				PathParams:   []ir.Param{{GoName: "path", JSONName: "path", Type: "string"}},
			},
		},
	}
	if err := matchInteractions(doc, cassette.Interactions{
		{
			Request:  cassette.Request{Method: "GET", URL: "https://pkg.go.dev/v1beta/package/github.com/google/go-cmp/cmp"},
			Response: cassette.Response{StatusCode: 200},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(doc.InteractionCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(doc.InteractionCalls))
	}
	want := `"github.com/google/go-cmp/cmp"`
	if doc.InteractionCalls[0].PathArgs[0] != want {
		t.Errorf("path arg = %q, want %q", doc.InteractionCalls[0].PathArgs[0], want)
	}
}
