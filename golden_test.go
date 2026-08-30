package codegen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	// The golden packages are built and run by TestCodegen_Golden rather than
	// by this module, so nothing else here imports what the generated code
	// needs. These keep those modules in go.mod and in vendor.
	_ "github.com/MarkRosemaker/openapi-enrich/cassette"
	_ "github.com/go-api-libs/api"
	_ "github.com/go-api-libs/api/server"
	_ "github.com/google/uuid"
)

// TestCodegen_Golden builds the generated packages and runs their tests.
//
// TestCodegen_TestData compares them as text, which proves the generator is
// deterministic and says nothing about whether what it emits is valid Go, let
// alone correct. The go tool skips anything under testdata, so it never found
// out: a parameter with no example rendered as the bare identifier null and
// every test stayed green. Naming a directory explicitly opts it back in.
//
// Running the tests rather than only type-checking them is what exercises the
// client against the interactions it was generated from -- the replay compares
// request headers exactly, so a wrong default reaches this and nothing else.
func TestCodegen_Golden(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join("testdata", entry.Name(), "pkg",
			strings.ReplaceAll(entry.Name(), "-", ""))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			out, err := exec.Command("go", "test", "./"+dir).CombinedOutput()
			if err != nil {
				t.Fatalf("generated code does not build or its tests fail:\n%s", out)
			}
		})
	}
}
