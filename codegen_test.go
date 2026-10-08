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
const (
	debugMode  = false
	production = true
)

//go:embed testdata
var testdata embed.FS

var genAll = config.Generate{
	Types:      true,
	Client:     true,
	ClientTest: true,
	Server:     true,
	JS:         true,
}

// TestCodegen_TestData generates each package in testdata from its specification and interactions, three times over,
// and must get the files there each time. golden holds every case the generator handles, once; auth and debug are
// what golden cannot be as well: a document whose every request sends credentials, and the one generated in debug
// mode. The files are edited by hand, never regenerated, so a change in what comes out is read before it is accepted.
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

			debug := debugMode || name == "debug"

			irDoc, err := ir.FromDocument(doc, name, "", production, debug)
			if err != nil {
				t.Fatalf("build IR: %v", err)
			}

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
						Debug:        debug,
						Spec:         doc,
						PackageName:  strings.ReplaceAll(name, "-", ""),
						OutputFs:     memFs,
						Interactions: ias,
						Generate:     genAll,
						Production:   production,
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

// compareBytes fails with the first line where the generated file differs from the one in testdata, which is to be
// edited by hand if the change is wanted.
func compareBytes(t *testing.T, expected, actual []byte, path string) {
	t.Helper()

	if bytes.Equal(expected, actual) {
		return
	}

	want, got := strings.Split(string(expected), "\n"), strings.Split(string(actual), "\n")

	i := 0
	for i < len(want) && i < len(got) && want[i] == got[i] {
		i++
	}

	line := func(lines []string) string {
		if i < len(lines) {
			return lines[i]
		}

		return "(end of file)"
	}

	t.Fatalf("%s:%d differs\nwant: %s\n got: %s", path, i+1, line(want), line(got))
}

// handWritten reports whether path is a file a golden directory carries but
// the generator does not produce. An error response type is passed to
// api.NewErrCustom, which takes an error, so the package author writes its
// Error method; what a good message reads like depends on the API, so the
// generator does not guess. A test the generator did not write is one too.
func handWritten(path string) bool {
	switch base := filepath.Base(path); base {
	case "error.go", "errors.go":
		return true
	default:
		return strings.HasSuffix(base, "_test.go") && !strings.HasSuffix(base, ".gen_test.go")
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
