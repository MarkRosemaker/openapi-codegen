package ir_test

import (
	"slices"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_FixedParams(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/v{major}/items": {
      "parameters": [
        {"name": "major", "in": "path", "required": true, "schema": {"type": "integer", "const": 2}},
        {"name": "Api-Version", "in": "header", "required": true, "schema": {"type": "string", "enum": ["2026-06-01"]}}
      ],
      "get": {
        "operationId": "listItems",
        "parameters": [
          {"name": "format", "in": "query", "required": true, "schema": {"type": "string", "const": "full view"}},
          {"name": "Example-Only", "in": "header", "required": true, "schema": {"type": "string", "example": "abc"}},
          {"name": "optional", "in": "query", "schema": {"type": "string", "const": "x"}},
          {"name": "choice", "in": "query", "required": true, "schema": {"type": "string", "enum": ["a", "b"]}}
        ],
        "responses": {"204": {"description": "none"}}
      }
    }
  }
}`))
	if err != nil {
		t.Fatal(err)
	}

	irDoc, err := ir.FromDocument(doc, "t", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	op := irDoc.Operations[0]

	names := func(ps ir.Params) []string {
		var out []string
		for _, p := range ps {
			out = append(out, p.JSONName)
		}

		return out
	}

	// a const or an only enum value of a required parameter is sent without being asked for; an example is not enough
	if got, want := names(op.FixedParams), []string{"major", "Api-Version", "format"}; !slices.Equal(got, want) {
		t.Errorf("FixedParams = %v, want %v", got, want)
	}

	if got := names(op.PathParams); len(got) != 0 {
		t.Errorf("PathParams = %v, want none", got)
	}

	if got, want := names(op.ParamsInStruct()), []string{"optional", "choice", "Example-Only"}; !slices.Equal(got, want) {
		t.Errorf("ParamsInStruct = %v, want %v", got, want)
	}

	if got, want := op.JoinPathArgs, []string{`"v" + "2"`, `"items"`}; !slices.Equal(got, want) {
		t.Errorf("JoinPathArgs = %v, want %v", got, want)
	}

	if got, want := names(op.FixedHeaders()), []string{"Api-Version"}; !slices.Equal(got, want) {
		t.Errorf("FixedHeaders = %v, want %v", got, want)
	}

	if got, want := op.FixedQuery(), "format=full+view"; got != want {
		t.Errorf("FixedQuery = %q, want %q", got, want)
	}

	if got, want := op.JSPathTemplate(), "/v2/items"; got != want {
		t.Errorf("JSPathTemplate = %q, want %q", got, want)
	}

	if got, want := op.JSPath(), "/v2/items?format=full+view"; got != want {
		t.Errorf("JSPath = %q, want %q", got, want)
	}
}
