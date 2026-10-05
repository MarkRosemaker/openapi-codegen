package render_test

import (
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-codegen/render"
)

func TestFiles_ClientTest_DecodingErrorOnlyForJSON(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/export": {"get": {"operationId": "export", "responses": {
      "200": {"description": "ok", "content": {"application/zip": {"schema": {"type": "string", "format": "binary"}}}}
    }}},
    "/readme": {"get": {"operationId": "readme", "responses": {
      "200": {"description": "ok", "content": {"text/plain": {"schema": {"type": "string"}}}}
    }}},
    "/item": {"get": {"operationId": "getItem", "responses": {
      "200": {"description": "ok", "content": {"application/vnd.api+json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}}
    }}}
  }
}`))
	if err != nil {
		t.Fatal(err)
	}

	irDoc, err := ir.FromDocument(doc, "t", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	files, err := render.Files(irDoc, genAll)
	if err != nil {
		t.Fatal(err)
	}

	var content string
	for _, f := range files {
		if f.Name == "client.gen_test.go" {
			content = string(f.Content)
		}
	}

	// one subtest per operation, in the order they are rendered
	subtests := map[string]string{}
	for _, op := range irDoc.Operations {
		start := strings.Index(content, `t.Run("`+op.Name+`"`)
		if start < 0 {
			t.Fatalf("no subtest for %s", op.Name)
		}

		subtests[op.Name] = content[start:]
		if end := strings.Index(subtests[op.Name][1:], "\n\tt.Run(\""); end >= 0 {
			subtests[op.Name] = subtests[op.Name][:end+1]
		}
	}

	for name, want := range map[string]bool{"Export": false, "Readme": false, "GetItem": true} {
		if got := strings.Contains(subtests[name], `"decoding error"`); got != want {
			t.Errorf("%s has a decoding error test: %v, want %v", name, got, want)
		}
	}

	// the body is sent as what the operation declares, or the client would reject it before decoding
	if !strings.Contains(subtests["GetItem"], `w.Header().Set("Content-Type", "application/vnd.api+json")`) {
		t.Error("GetItem's decoding error test does not send its own media type")
	}
}
