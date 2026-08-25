package main

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/MarkRosemaker/fsutil/osutil"
	"github.com/MarkRosemaker/jsonutil"
	"github.com/MarkRosemaker/openapi"
	codegen "github.com/MarkRosemaker/openapi-codegen"
	"github.com/MarkRosemaker/openapi-codegen/config"
	"github.com/MarkRosemaker/openapi-codegen/ir"
	"github.com/MarkRosemaker/openapi-codegen/render"
)

var genAll = config.Generate{
	Types:      true,
	Client:     true,
	ClientTest: true,
	Server:     true,
	JS:         true,
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if err := copyPreviousStep(); err != nil {
		return err
	}

	petDoc := &ir.Document{}
	petDir := filepath.Join("render", "testdata", "pet")
	petIRPath := filepath.Join(petDir, "ir.json")
	{
		// build pet doc
		f, err := os.Open(petIRPath)
		if err != nil {
			return err
		}
		defer f.Close() //nolint:errcheck

		if err := json.UnmarshalRead(f, petDoc); err != nil {
			return err
		}

		files, err := render.Files(petDoc, genAll)
		if err != nil {
			return err
		}

		dir := filepath.Join(petDir, "golden")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		for _, f := range files {
			path := filepath.Join(dir, f.Name)
			if err := os.WriteFile(path, f.Content, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
	}

	{
		// rewrite pet doc
		f, err := os.Create(petIRPath)
		if err != nil {
			return err
		}
		defer f.Close() //nolint:errcheck

		if err := json.MarshalWrite(f, petDoc, jsontext.Multiline(true)); err != nil {
			return err
		}
	}

	entries, err := os.ReadDir("testdata")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		doc, err := openapi.LoadFromFile(filepath.Join("testdata", name, "api", "openapi.json"))
		if err != nil {
			return err
		}

		irDoc, err := ir.FromDocument(doc, name, "")
		if err != nil {
			return fmt.Errorf("build IR: %w", err)
		}

		// Mirrors the layout the CLI produces -- api/ beside pkg/<package> --
		// so the generated code is what a real project gets, the hardcoded
		// cassette path in the generated test included.
		outDir := filepath.Join("testdata", name, "pkg", strings.ReplaceAll(name, "-", ""))

		if err := os.MkdirAll(outDir, 0o700); err != nil {
			return err
		}

		if err := jsonutil.WriteFile(filepath.Join(outDir, "ir.json"), irDoc,
			jsontext.Multiline(true)); err != nil {
			return err
		}

		iaPath := filepath.Join("testdata", name, "api", "interactions.json")
		if _, err := os.Stat(iaPath); errors.Is(err, fs.ErrNotExist) {
			iaPath = ""
		} else if err != nil {
			return err
		}

		if err := codegen.Generate(codegen.Config{
			Spec:             doc,
			PackageName:      strings.ReplaceAll(entry.Name(), "-", ""),
			OutputDir:        outDir,
			InteractionsPath: iaPath,
			Generate:         genAll,
		}); err != nil {
			return err
		}
	}

	return nil
}

func copyPreviousStep() error {
	const enrichDir = "../openapi-enrich/testdata"
	const compressDir = "../openapi-compress/testdata"

	for srcDir, names := range map[string][2]string{
		enrichDir:   {"interactions.json", "interactions.json"},
		compressDir: {"golden.json", "openapi.json"},
	} {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}

			return fmt.Errorf("reading %q: %w", srcDir, err)
		}

		for _, e := range entries {
			if err := osutil.Copy(
				filepath.Join(srcDir, e.Name(), names[0]),
				filepath.Join("testdata", e.Name(), "api", names[1]),
			); err != nil {
				return fmt.Errorf("copying %q: %w", srcDir, err)
			}
		}

	}

	return nil
}
