- **types** — structs, enums, and type aliases for all referenced schemas
- **client** — typed HTTP client with per-operation methods
- **server** — `http.Handler`-based server scaffold
- **tests** — round-trip and cassette-backed tests, generated from recorded traffic
- **JavaScript client** — optional `api.js` alongside the Go output

Identifiers are sanitized into valid, idiomatic Go: leading digits are spelled out,
punctuation is stripped, acronyms are preserved, and names that would collide with
the `error` interface or a Go keyword are renamed.

How the specification maps onto Go:

- **Unions** — a `oneOf` or `anyOf` becomes a struct with one field per
  alternative, exactly one (`oneOf`) or at least one (`anyOf`) of them set after
  decoding. A field is a pointer, nil until set, unless its zero value already
  says it is not set: nil for a slice or a map, `""` for a string. An alternative that is only `null` needs no field. Where a member tells
  the alternatives apart (the `discriminator`'s `propertyName`, or a member each
  alternative fixes to a string of its own, such as Notion's `type`), that member
  names the alternative. When it comes first, the alternative it names decodes
  each further member as it is read, without reading the whole value first;
  when it does not, the value is read whole and then decoded the same way. An
  unknown value, or a value without that member, is an error. An alternative that is a union of
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
- **Debug mode** — with `-debug`, a client given `WithDebug` records each
  response it fails to decode to `api/interactions.json`, for `openapi-enrich` to
  learn from. Nothing the specification leaves open decodes into `any` then: the
  empty schema, an array without `items`, a free-form object and `not` alone
  become `struct{}`, so any value in them fails and is recorded.
- **Fields** — a field is a pointer only where its zero value must be told apart
  from something else: from leaving the field out, if it is optional, or from
  null, if it is nullable. That is a boolean, a number that may be 0, or an object
  that requires nothing, so `{}` says something; nil then leaves the field out,
  and a pointer to the zero value sends it. A zero value the specification makes
  the default, or rules out with a bound or an enum, needs no pointer. A string,
  a time, a slice, a map or a union is never a pointer: its zero value leaves it
  out. A struct that would contain itself refers to itself through a pointer.
- **Omitting** — an optional field is tagged `omitzero`, so it is left out while
  it holds its zero value: nil for a pointer, a slice or a map, so an empty one
  is still sent. A required field is always sent, `""`, `0` and `false`
  included.
- **Null** — a schema that is only ever `null` is `*struct{}`, and "X or null"
  is `X`, a pointer to `X` only by the rule for fields above, so that null and
  the zero value can differ.
- **Fixed parameters** — a required parameter that can take only one value, its
  `const` or the only value of its `enum`, is sent by the client itself: in the
  path, the query or the headers, such as Notion's `Notion-Version`. The caller
  never passes it. An `example` alone does not fix a value.
- **Query parameters** — an array is sent as one value per element (form style,
  exploded). "X or an array of X" is sent as the array, a union of strings as a
  string, and `null` is dropped, since a query string cannot carry it.
- **Authentication** — each operation sends the credential its own `security`
  names, else the document's: a bearer token or basic auth, read from the
  environment. Where a document uses both, `NewClient` requires at least one, and
  each operation fails before sending if its own is missing.
- **Success responses** — an operation returns its success body as `*T`, or as
  `T` where `T` is already nilable: a slice, a map, or a named type of either. An
  operation whose success body is an empty object returns just `error`, and the
  server writes `{}`. Its body is read only in debug mode, where it is decoded so
  that anything in it fails loudly and is recorded.
- **Error responses** — an error body's type is returned wrapped in
  `api.Error`, so it needs an `Error() string` method. The generator does not
  write one, since a good message depends on the API: add it by hand beside the
  generated code.
- **Names** — a component's Go name is its `x-go-name`, or its key with any
  character a Go identifier cannot hold removed (`Keypoint-Input` →
  `KeypointInput`). Every type is exported: a key that does not begin with an
  upper-case letter is renamed in Go style first (`idRequest` → `IDRequest`,
  `error_api_400` → `ErrorAPI400`), with every reference to it. A name another
  component already holds is numbered (`date` beside `Date` → `Date2`); set
  `x-go-name` to choose a better one. A component that is only a `$ref`, only
  `null` or the empty schema declares no type; references to it use what it
  stands for.
