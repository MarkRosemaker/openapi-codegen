- **types** — structs, enums, and type aliases for all referenced schemas
- **client** — typed HTTP client with per-operation methods
- **server** — `http.Handler`-based server scaffold
- **tests** — round-trip and cassette-backed tests, generated from recorded traffic
- **JavaScript client** — optional `api.js` alongside the Go output

Identifiers are sanitized into valid, idiomatic Go: leading digits are spelled out,
punctuation is stripped, acronyms are preserved, and names that would collide with
the `error` interface or a Go keyword are renamed.

How the specification maps onto Go:

- **Unions** — a `oneOf` or `anyOf` becomes a struct with one pointer field per
  alternative, exactly one (`oneOf`) or at least one (`anyOf`) of them set after
  decoding. An alternative that is only `null` needs no field.
- **Null** — a schema that is only ever `null` is `*struct{}`, and "X or null"
  is `*X`, or plain `X` where `X` is already nilable (a slice, a map, `any`).
- **Query parameters** — an array is sent as one value per element (form style,
  exploded). "X or an array of X" is sent as the array, a union of strings as a
  string, and `null` is dropped, since a query string cannot carry it.
- **Authentication** — each operation sends the credential its own `security`
  names, else the document's: a bearer token or basic auth, read from the
  environment. `NewClient` requires only the default one.
- **Error responses** — an error body's type is returned wrapped in
  `api.Error`, so it needs an `Error() string` method. The generator does not
  write one, since a good message depends on the API: add it by hand beside the
  generated code.
- **Names** — a component's Go name is its `x-go-name`, or its key with any
  character a Go identifier cannot hold removed (`Keypoint-Input` →
  `KeypointInput`). A component that is only a `$ref`, only `null` or the empty
  schema declares no type; references to it use what it stands for.
