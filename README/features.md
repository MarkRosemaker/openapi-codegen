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
  decoding. An alternative that is only `null` needs no field. Where a member tells
  the alternatives apart (the `discriminator`'s `propertyName`, or a member each
  alternative fixes to a string of its own, such as Notion's `type`), that member
  must come first: decoding reads it, and the alternative it names decodes each
  further member as it is read, without reading the whole value first. An unknown
  value, or a different first member, is an error. An alternative that is a union of
  its own counts by its alternatives, however deep, so its leaves are chosen the same
  way. Encoding writes the discriminator first, with the value of the alternative
  set, and refuses a different one. Otherwise each alternative is tried in turn.
- **allOf** — each part referenced by this schema alone is folded into its fields;
  a part other schemas share stays an embedded type. A union among the parts is a
  field of its own, decoded by the struct's methods: the fields and the chosen
  alternative each take the members they declare, chosen by the discriminator or by
  which members are present, and a member none of them declares is an error.
- **Strictness** — decoding fails as soon as the input differs from what the
  specification allows, so that an incomplete specification shows itself. The
  generated methods report it as `encoding/json` does, with a `*json.SemanticError`
  locating it in the input; an unknown member wraps `json.ErrUnknownName`. A case
  that could be supported but has no real example yet, such as an `allOf` of two
  unions, is generated with methods that return an "unimplemented" error.
- **Null** — a schema that is only ever `null` is `*struct{}`, and "X or null"
  is `*X`, or plain `X` where `X` is already nilable (a slice, a map, `any`).
- **Query parameters** — an array is sent as one value per element (form style,
  exploded). "X or an array of X" is sent as the array, a union of strings as a
  string, and `null` is dropped, since a query string cannot carry it.
- **Authentication** — each operation sends the credential its own `security`
  names, else the document's: a bearer token or basic auth, read from the
  environment. Where a document uses both, `NewClient` requires at least one, and
  each operation fails before sending if its own is missing.
- **Error responses** — an error body's type is returned wrapped in
  `api.Error`, so it needs an `Error() string` method. The generator does not
  write one, since a good message depends on the API: add it by hand beside the
  generated code.
- **Names** — a component's Go name is its `x-go-name`, or its key with any
  character a Go identifier cannot hold removed (`Keypoint-Input` →
  `KeypointInput`). A component that is only a `$ref`, only `null` or the empty
  schema declares no type; references to it use what it stands for.
