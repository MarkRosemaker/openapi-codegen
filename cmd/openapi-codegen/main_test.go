package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const spec = `{
  "openapi": "3.1.0",
  "info": {"title": "Pets", "version": "1.0.0"},
  "servers": [{"url": "https://pets.example.com"}],
  "paths": {
    "/pets/{id}": {
      "get": {
        "operationId": "getPet",
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {
          "200": {
            "description": "OK",
            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Pet"}}}
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Pet": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}
    }
  }
}`

// project is a working directory named my-API holding the spec at the default path.
func project(t *testing.T) string {
	t.Helper()

	wd := filepath.Join(t.TempDir(), "my-API")
	if err := os.MkdirAll(filepath.Join(wd, "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(wd, "api", "openapi.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(wd)

	return wd
}

// generated lists the files in dir.
func generated(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}

	return names
}

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		dir  string // where the files go, relative to the working directory
		want []string
	}{
		// the package and output directory default to the directory's name, lower case and without dashes
		{"client", []string{"-client"}, "pkg/myapi", []string{"client.gen.go", "client.gen_test.go", "types.gen.go"}},
		{"server", []string{"-server"}, "pkg/myapi", []string{"server.gen.go", "types.gen.go"}},
		{"js alone needs no types", []string{"-js", "-out", "web"}, "web", []string{"api.js"}},
		{"all", []string{"-client", "-server", "-js", "-out", "gen"}, "gen", []string{
			"api.js", "client.gen.go", "client.gen_test.go", "server.gen.go", "types.gen.go",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wd := project(t)

			var stderr bytes.Buffer
			if err := run(tc.args, wd, &stderr); err != nil {
				t.Fatalf("run: %v\n%s", err, stderr.String())
			}

			if got := generated(t, filepath.Join(wd, tc.dir)); strings.Join(got, " ") != strings.Join(tc.want, " ") {
				t.Errorf("generated %v, want %v", got, tc.want)
			}

			// debug mode is only on when asked for
			if client, err := os.ReadFile(filepath.Join(wd, tc.dir, "client.gen.go")); err == nil && bytes.Contains(client, []byte("WithDebug")) {
				t.Error("client.gen.go is in debug mode")
			}
		})
	}
}

func TestRun_Flags(t *testing.T) {
	wd := project(t)

	if err := os.Rename(filepath.Join(wd, "api"), filepath.Join(wd, "spec")); err != nil {
		t.Fatal(err)
	}

	args := []string{"-client", "-spec", "spec/openapi.json", "-pkg", "pets", "-agent", "pets-cli/2.0", "-debug"}
	if err := run(args, wd, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	// the output directory follows the package
	client, err := os.ReadFile(filepath.Join(wd, "pkg", "pets", "client.gen.go"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"package pets", `"pets-cli/2.0"`, "WithDebug"} {
		if !bytes.Contains(client, []byte(want)) {
			t.Errorf("client.gen.go lacks %s", want)
		}
	}
}

func TestRun_Errors(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string // in the error
	}{
		{"nothing to generate", nil, "either -client, -server, or -js are required"},
		{"no spec", []string{"-client", "-spec", "missing.json"}, "missing.json"},
		{"unknown flag", []string{"-types"}, "flag provided but not defined: -types"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wd := project(t)

			var stderr bytes.Buffer

			err := run(tc.args, wd, &stderr)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got error %v, want one containing %q", err, tc.want)
			}

			if !strings.Contains(stderr.String(), "Usage of openapi-codegen") {
				t.Errorf("no usage on stderr:\n%s", stderr.String())
			}

			if _, err := os.Stat(filepath.Join(wd, "pkg")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("wrote output anyway: %v", err)
			}
		})
	}
}

func TestRun_Help(t *testing.T) {
	var stderr bytes.Buffer
	if err := run([]string{"-h"}, t.TempDir(), &stderr); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp", err)
	}

	if !strings.Contains(stderr.String(), "-client") {
		t.Errorf("usage lists no flags:\n%s", stderr.String())
	}
}
