package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_DebugLeavesNothingOpen(t *testing.T) {
	t.Parallel()

	const spec = `{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {},
  "components": {"schemas": {
    "Page": {
      "type": "object",
      "required": ["meta"],
      "properties": {
        "meta": {},
        "rows": {"type": "array"},
        "extra": {"type": "object", "additionalProperties": true},
        "labels": {"type": "object", "additionalProperties": {}},
        "counts": {"type": "object", "additionalProperties": {"type": "integer"}},
        "last": {"anyOf": [{"type": "object", "additionalProperties": true}, {"type": "null"}]},
        "parent": {"not": {}}
      }
    }
  }}
}`

	for debug, want := range map[bool]map[string]string{
		false: {
			"meta":   "any",
			"rows":   "[]any",
			"extra":  "map[string]any",
			"labels": "map[string]any",
			"counts": "map[string]int",
			"last":   "map[string]any",
			"parent": "any",
		},
		// what the specification leaves open fails to decode once it holds anything
		true: {
			"meta":   "struct{}",
			"rows":   "[]struct{}",
			"extra":  "*struct{}",
			"labels": "*struct{}",
			"counts": "map[string]int",
			"last":   "*struct{}",
			"parent": "*struct{}",
		},
	} {
		doc, err := openapi.LoadFromDataJSON([]byte(spec))
		if err != nil {
			t.Fatal(err)
		}

		irDoc, err := ir.FromDocument(doc, "t", "", false, debug)
		if err != nil {
			t.Fatal(err)
		}

		if irDoc.Debug != debug {
			t.Errorf("debug %v: Document.Debug = %v", debug, irDoc.Debug)
		}

		for _, f := range schemaNamed(t, irDoc, "Page").Fields {
			if f.Type != want[f.JSONName] {
				t.Errorf("debug %v: %s is %s, want %s", debug, f.JSONName, f.Type, want[f.JSONName])
			}
		}
	}
}
