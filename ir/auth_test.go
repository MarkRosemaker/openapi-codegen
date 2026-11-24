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

// TestFromDocument_APIKey reads an apiKey security scheme, in a header, a query parameter or a cookie, as the key the
// client sends there on each operation that requires it, read from the environment variable named after the API.
// Where a spec also has an X-Api-Key header parameter, the scheme wins: a client has one API key.
func TestFromDocument_APIKey(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"header", "query", "cookie"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			doc, err := openapi.LoadFromDataJSON([]byte(`{
  "openapi": "3.1.0",
  "info": {"title": "Finnhub API", "version": "1"},
  "security": [{"api_key": []}],
  "paths": {
    "/quote": {"get": {"operationId": "quote", "parameters": [{"name": "X-Api-Key", "in": "header", "required": true, "schema": {"type": "string"}}], "responses": {"204": {"description": "ok"}}}},
    "/health": {"get": {"operationId": "health", "security": [], "responses": {"204": {"description": "ok"}}}}
  },
  "components": {"securitySchemes": {"api_key": {"type": "apiKey", "in": "` + in + `", "name": "X-Finnhub-Token"}}}
}`))
			if err != nil {
				t.Fatal(err)
			}

			irDoc, err := ir.FromDocument(doc, "finnhub", "", false, false)
			if err != nil {
				t.Fatal(err)
			}

			if want := (ir.APIKey{In: in, Name: "X-Finnhub-Token", EnvName: "FINNHUB_API_KEY"}); irDoc.Auth.APIKey != want {
				t.Errorf("API key is %+v, want %+v", irDoc.Auth.APIKey, want)
			}

			if irDoc.Auth.Default != ir.AuthAPIKey {
				t.Errorf("default auth is %q, want apiKey", irDoc.Auth.Default)
			}

			for _, op := range irDoc.Operations {
				if want := map[string]ir.AuthScheme{"Quote": ir.AuthAPIKey}[op.Name]; op.Auth != want {
					t.Errorf("%s sends %q, want %q", op.Name, op.Auth, want)
				}
			}

			if p := irDoc.APIKey(); p != nil {
				t.Errorf("the X-Api-Key parameter is an API key too: %+v", p)
			}
		})
	}
}
