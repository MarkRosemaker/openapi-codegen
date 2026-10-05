package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_StreamsBinarySuccess(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "t", "version": "1"},
  "paths": {
    "/export": {"get": {"operationId": "export", "responses": {
      "200": {"description": "ok", "content": {"application/zip": {"schema": {"type": "string", "format": "binary"}}}},
      "400": {"description": "bad", "content": {"application/pdf": {"schema": {"type": "string", "format": "binary"}}}}
    }}},
    "/readme": {"get": {"operationId": "readme", "responses": {
      "200": {"description": "ok", "content": {"text/plain": {"schema": {"type": "string"}}}}
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

	ops := map[string]ir.Operation{}
	for _, op := range irDoc.Operations {
		ops[op.Name] = op
	}

	// a zip is handed over as it arrives; an error body is read whole, as is text
	export := ops["Export"]
	if !export.StreamSuccess || export.SuccessReturn.String() != "io.ReadCloser" || export.SuccessReturn.Nilable() != "io.ReadCloser" {
		t.Errorf("export: stream %v, returns %v", export.StreamSuccess, export.SuccessReturn)
	}

	for _, r := range export.Responses {
		if r.IsSuccess != r.IsStream || !r.IsSuccess && !r.IsRawBytes {
			t.Errorf("export %s: stream %v, raw bytes %v", r.StatusCode, r.IsStream, r.IsRawBytes)
		}
	}

	if readme := ops["Readme"]; readme.StreamSuccess || !readme.RawBytesSuccess {
		t.Errorf("readme: stream %v, raw bytes %v; want text read whole", readme.StreamSuccess, readme.RawBytesSuccess)
	}
}
