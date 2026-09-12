- **types** — structs, enums, and type aliases for all referenced schemas
- **client** — typed HTTP client with per-operation methods
- **server** — `http.Handler`-based server scaffold
- **tests** — round-trip and cassette-backed tests, generated from recorded traffic
- **JavaScript client** — optional `api.js` alongside the Go output

Identifiers are sanitized into valid, idiomatic Go: leading digits are spelled out,
punctuation is stripped, acronyms are preserved, and names that would collide with
the `error` interface or a Go keyword are renamed.
