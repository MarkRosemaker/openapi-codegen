package codegen_test

import (
	"bytes"
	"embed"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarkRosemaker/openapi"
	codegen "github.com/MarkRosemaker/openapi-codegen"
	"github.com/MarkRosemaker/openapi-codegen/config"
	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-enrich/cassette"
	"github.com/spf13/afero"
)

// Whether to test in debug mode.
const debugMode = false

//go:embed testdata
var testdata embed.FS

var genAll = config.Generate{
	Types:      true,
	Client:     true,
	ClientTest: true,
	Server:     true,
	JS:         true,
}

func TestCodegen_TestData(t *testing.T) {
	entries, err := testdata.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range entries {
		t.Run(tc.Name(), func(t *testing.T) {
			name := tc.Name()

			f, err := testdata.Open(filepath.Join("testdata", name, "api", "openapi.json"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close() //nolint

			doc, err := openapi.LoadFromReader(f)
			if err != nil {
				t.Fatal(err)
			}

			irDoc, err := ir.FromDocument(doc, name, "")
			if err != nil {
				t.Fatalf("build IR: %v", err)
			}

			irDoc.Debug = debugMode

			iasPath := filepath.Join("testdata", name, "api", "interactions.json")
			ias, err := cassette.InteractionsReadFile(iasPath)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				t.Fatal(err)
			}

			for it := range 3 {
				t.Run(fmt.Sprintf("iteration %d", it+1), func(t *testing.T) {
					memFs := afero.NewMemMapFs()

					writeJSON(t, memFs, "ir.json", irDoc)

					if err := codegen.Generate(codegen.Config{
						Debug:        debugMode,
						Spec:         doc,
						PackageName:  strings.ReplaceAll(name, "-", ""),
						OutputFs:     memFs,
						Interactions: ias,
						Generate:     genAll,
					}); err != nil {
						t.Fatal(err)
					}

					wantFs := afero.NewBasePathFs(afero.NewOsFs(), filepath.Join("testdata", name, "pkg", strings.ReplaceAll(name, "-", "")))
					compareFs(t, wantFs, memFs)
				})
			}
		})
	}
}

// compareBytes prints a compact diff of two byte slices
func compareBytes(t *testing.T, expected, actual []byte, path string) {
	t.Helper()

	if bytes.Equal(expected, actual) {
		return
	}

	// Find first difference
	i := 0
	for i < len(expected) && i < len(actual) && expected[i] == actual[i] {
		i++
	}

	t.Errorf("\n┌─ Diff in %s at offset %d\n│ Expected: %q\n│ Actual:   %q\n└─ %s",
		path, i, expected[i:min(len(expected), i+20)], actual[i:min(len(actual), i+20)],
		func() string {
			if len(expected) != len(actual) {
				return fmt.Sprintf("length %d vs %d", len(expected), len(actual))
			}
			return fmt.Sprintf("0x%02x vs 0x%02x", expected[i], actual[i])
		}())
}

// handWritten reports whether path is a file a golden directory carries but
// the generator does not produce. An error response type is passed to
// api.NewErrCustom, which takes an error, so the package author writes its
// Error method; what a good message reads like depends on the API, so the
// generator does not guess.
func handWritten(path string) bool {
	switch filepath.Base(path) {
	case "error.go", "errors.go":
		return true
	default:
		return false
	}
}

func compareFs(t *testing.T, expected, actual afero.Fs) {
	t.Helper()

	if err := afero.Walk(expected, "", func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if handWritten(path) {
			return nil
		}

		want, err := afero.ReadFile(expected, path)
		if err != nil {
			return err
		}

		got, err := afero.ReadFile(actual, path)
		if err != nil {
			return fmt.Errorf("in actual FS: %w", err)
		}

		compareBytes(t, want, got, path)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := afero.Walk(actual, "", func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if ok, err := afero.Exists(expected, path); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("generated additional file %q", path)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, fsys afero.Fs, path string, in any) {
	f, err := fsys.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck

	if ias, ok := in.(cassette.Interactions); ok {
		if err := ias.MarshalWrite(f); err != nil {
			t.Fatal(err)
		}
	} else if err := json.MarshalWrite(f, in, jsontext.Multiline(true)); err != nil {
		t.Fatal(err)
	}
}
