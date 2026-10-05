package ir_test

import (
	"testing"

	"github.com/MarkRosemaker/openapi"
	"github.com/MarkRosemaker/openapi-codegen/ir"
)

func TestFromDocument_AuthPerOperation(t *testing.T) {
	t.Parallel()

	doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "Pets", "version": "1"},
  "security": [{"bearerAuth": []}],
  "paths": {
    "/pets": {"get": {"operationId": "listPets", "responses": {"204": {"description": "ok"}}}},
    "/token": {"post": {"operationId": "createToken", "security": [{"basicAuth": []}], "responses": {"204": {"description": "ok"}}}},
    "/health": {"get": {"operationId": "health", "security": [], "responses": {"204": {"description": "ok"}}}},
    "/feed": {"get": {"operationId": "getFeed", "security": [{}, {"bearerAuth": []}], "responses": {"204": {"description": "ok"}}}}
  },
  "components": {
    "securitySchemes": {
      "basicAuth": {"type": "http", "scheme": "basic"},
      "bearerAuth": {"type": "http", "scheme": "bearer"}
    }
  }
}`))
	if err != nil {
		t.Fatal(err)
	}

	irDoc, err := ir.FromDocument(doc, "pets", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if irDoc.Auth.Default != ir.AuthBearer {
		t.Errorf("default auth is %q, want bearer", irDoc.Auth.Default)
	}

	want := map[string]ir.AuthScheme{"ListPets": ir.AuthBearer, "CreateToken": ir.AuthBasic, "Health": "", "GetFeed": ir.AuthBearer}
	for _, op := range irDoc.Operations {
		if op.Auth != want[op.Name] {
			t.Errorf("%s sends %q, want %q", op.Name, op.Auth, want[op.Name])
		}

		// optional credentials are sent all the same, and only GetFeed's are optional
		if op.AuthOptional != (op.Name == "GetFeed") {
			t.Errorf("%s: credentials optional is %v", op.Name, op.AuthOptional)
		}
	}
}
