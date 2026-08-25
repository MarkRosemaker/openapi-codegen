package codegen_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/MarkRosemaker/openapi"
	codegen "github.com/MarkRosemaker/openapi-codegen"
	"github.com/MarkRosemaker/openapi-codegen/config"
)

func TestGenerate_HappyPath(t *testing.T) {
	outDir := t.TempDir()

	cfg := codegen.Config{
		SpecPath:    filepath.Join("testdata", "simple", "api", "openapi.json"),
		OutputDir:   outDir,
		PackageName: "simple",
		UserAgent:   "simple/1.0",
	}

	if err := codegen.Generate(cfg); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Verify expected files were written.
	wantFiles := []string{"types.gen.go", "client.gen.go", "client.gen_test.go"}
	for _, name := range wantFiles {
		path := filepath.Join(outDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file %s not found: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Fatalf("file %s is empty", name)
		}
	}

	// Spot-check types.go content.
	types, err := os.ReadFile(filepath.Join(outDir, "types.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(types)
	if !containsStr(content, "package simple") {
		t.Error("types.go: missing package declaration")
	}
	if !containsStr(content, "type Item struct") {
		t.Fatalf("types.go: missing Item struct; content:\n%s", content)
	}
}

func TestGenerate_SelectiveFiles(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cfg       config.Generate
		wantFiles []string
		noFiles   []string
	}{
		{
			name:      "only types",
			cfg:       config.Generate{Types: true},
			wantFiles: []string{"types.gen.go"},
			noFiles:   []string{"client.gen.go", "client.gen_test.go", "server.gen.go"},
		},
		{
			name:      "client and server",
			cfg:       config.Generate{Client: true, Server: true},
			wantFiles: []string{"client.gen.go", "server.gen.go"},
			noFiles:   []string{"types.gen.go", "client.gen_test.go"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outDir := t.TempDir()

			if err := codegen.Generate(codegen.Config{
				SpecPath:    filepath.Join("testdata", "simple", "api", "openapi.json"),
				OutputDir:   outDir,
				PackageName: "simple",
				Generate:    tc.cfg,
			}); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			for _, name := range tc.wantFiles {
				if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
					t.Errorf("expected %s to be generated: %v", name, err)
				}
			}

			for _, name := range tc.noFiles {
				if _, err := os.Stat(filepath.Join(outDir, name)); err == nil {
					t.Errorf("expected %s NOT to be generated", name)
				}
			}
		})
	}
}

func TestGenerate_EmptySpecPath(t *testing.T) {
	err := codegen.Generate(codegen.Config{PackageName: "pkg"})
	if err == nil {
		t.Fatal("expected error for empty SpecPath")
	}
}

func TestGenerate_EmptyPackageName(t *testing.T) {
	err := codegen.Generate(codegen.Config{
		Spec: &openapi.Document{
			OpenAPI: "3.1.0",
			Info: &openapi.Info{
				Title:   "My API",
				Version: "1.0",
			},
			Paths: openapi.Paths{
				"/foo": &openapi.PathItem{},
			},
		},
	})
	if err == nil {
		t.Fatal("expected error for empty PackageName")
	} else if got, want := err.Error(), "PackageName is required"; got != want {
		t.Fatalf("error=%q, want=%q", got, want)
	}
}

func TestGenerate_NoSpec(t *testing.T) {
	err := codegen.Generate(codegen.Config{
		OutputDir:   t.TempDir(),
		PackageName: "pkg",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent spec")
	} else if got, want := err.Error(), "SpecPath is required"; got != want {
		t.Fatalf("error=%q, want=%q", got, want)
	}
}

func TestGenerate_BadSpec(t *testing.T) {
	err := codegen.Generate(codegen.Config{
		Spec:        &openapi.Document{},
		PackageName: "pkg",
		OutputDir:   t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for nonexistent spec")
	} else if got, want := err.Error(), "invalid spec given: openapi is required"; got != want {
		t.Fatalf("error=%q, want=%q", got, want)
	}
}

func TestGenerate_IRError(t *testing.T) {
	// Spec with a path operation missing operationId → FromOperation returns an error.
	err := codegen.Generate(codegen.Config{
		Spec: &openapi.Document{
			OpenAPI: "3.0.3",
			Info: &openapi.Info{
				Title:   "no-op-id",
				Version: "1.0.0",
			},
			Paths: openapi.Paths{
				"/test": &openapi.PathItem{
					Get: &openapi.Operation{},
				},
			},
		},
		OutputDir:   t.TempDir(),
		PackageName: "pkg",
	})
	if err == nil {
		t.Fatal("expected error for spec with missing operationId")
	} else if got, want := err.Error(), "build IR: paths: GET /test: operationId is required"; got != want {
		t.Fatalf("error=%q, want=%q", got, want)
	}
}

func TestGenerate_MkdirError(t *testing.T) {
	// Use a regular file as OutputDir so MkdirAll fails.
	tmpFile, err := os.CreateTemp("", "codegen-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:errcheck
	_ = tmpFile.Close()

	if err := codegen.Generate(codegen.Config{
		SpecPath:    filepath.Join("testdata", "simple", "api", "openapi.json"),
		OutputDir:   filepath.Join(tmpFile.Name(), "subdir"),
		PackageName: "simple",
	}); err == nil {
		t.Fatal("expected error when output dir cannot be created")
	} else if pathErr, ok := errors.AsType[*fs.PathError](err); !ok {
		t.Fatalf("error=%v, want=%T", err, pathErr)
	} else if got, want := pathErr.Op, "mkdir"; got != want {
		t.Fatalf("(*fs.PathError).Op=%v, want=%q", got, want)
	}
}

func TestGenerate_WriteFileError(t *testing.T) {
	outDir := t.TempDir()
	// Pre-create a directory where types.go would be written — os.WriteFile will fail.
	if err := os.Mkdir(filepath.Join(outDir, "types.gen.go"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := codegen.Generate(codegen.Config{
		SpecPath:    filepath.Join("testdata", "simple", "api", "openapi.json"),
		OutputDir:   outDir,
		PackageName: "simple",
	}); err == nil {
		t.Fatal("expected error when output path is a directory")
	} else if pathErr, ok := errors.AsType[*fs.PathError](err); !ok {
		t.Fatalf("error=%v, want=%T", err, pathErr)
	} else if got, want := pathErr.Op, "open"; got != want {
		t.Fatalf("(*fs.PathError).Op=%v, want=%q", got, want)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && indexStr(s, sub) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
