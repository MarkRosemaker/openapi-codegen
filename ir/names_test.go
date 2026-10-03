package ir_test

import (
	"slices"
	"testing"
)

func TestFromDocument_ExportsComponentNames(t *testing.T) {
	t.Parallel()

	doc := loadSchemas(t, `{
		"Page": {
			"type": "object",
			"properties": {
				"id": {"$ref": "#/components/schemas/idRequest"},
				"when": {"$ref": "#/components/schemas/date"},
				"error": {"$ref": "#/components/schemas/error_api_400"},
				"custom": {"$ref": "#/components/schemas/custom"}
			}
		},
		"idRequest": {"type": "object", "properties": {"x": {"type": "string"}}},
		"Date": {"type": "object", "properties": {"date": {"$ref": "#/components/schemas/date"}}},
		"date": {"type": "object", "properties": {"start": {"type": "string"}}},
		"error_api_400": {"type": "object", "properties": {"code": {"type": "string"}}},
		"custom": {"type": "object", "x-go-name": "Chosen", "properties": {"y": {"type": "string"}}},
		"URL2": {"type": "object", "properties": {"href": {"type": "string"}}}
	}`)

	var names []string
	for _, s := range doc.Schemas {
		names = append(names, s.Name)
	}

	// an exported name stays as it is, even one Go-style casing would change; the second Date is numbered
	for _, want := range []string{"Page", "IDRequest", "Date", "Date2", "ErrorAPI400", "Chosen", "URL2"} {
		if !slices.Contains(names, want) {
			t.Errorf("no schema %s in %v", want, names)
		}
	}

	types := map[string]string{}
	for _, f := range schemaNamed(t, doc, "Page").Fields {
		types[f.JSONName] = f.Type
	}

	for name, want := range map[string]string{"id": "*IDRequest", "when": "*Date2", "error": "*ErrorAPI400", "custom": "*Chosen"} {
		if types[name] != want {
			t.Errorf("Page.%s is %s, want %s", name, types[name], want)
		}
	}
}
