```bash
go get -tool github.com/MarkRosemaker/openapi-codegen/cmd/openapi-codegen
```

or

```bash
go get github.com/MarkRosemaker/openapi-codegen
```


```sh
openapi-codegen -spec openapi.json -out ./gen -pkg mypkg -client
```

At least one of `-client`, `-server`, or `-js` is required. Types are generated
automatically whenever a client or server is, and client tests whenever a client is.

| Flag | Default | Purpose |
|---|---|---|
| `-spec` | `api/openapi.json` | Path to the OpenAPI specification |
| `-out` | `pkg/<package>` | Output directory for generated files |
| `-pkg` | directory name | Go package name |
| `-agent` | — | `User-Agent` string for the generated client |
| `-client` | `false` | Generate `client.gen.go` and `client.gen_test.go` |
| `-server` | `false` | Generate `server.gen.go` |
| `-js` | `false` | Generate `api.js` |

It can also be used as a library:

```go
import "github.com/MarkRosemaker/openapi-codegen"

err := codegen.Generate(codegen.Config{
    SpecPath:    "api/openapi.json",
    OutputDir:   "pkg/mypkg",
    PackageName: "mypkg",
})
```

`Config` accepts an already-parsed `*openapi.Document` instead of `SpecPath`, and an
`afero.Fs` instead of `OutputDir`, so generation can run entirely in memory.
