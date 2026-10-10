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
  set, and refuses a different one. Otherwise each alternative is tried in turn,
  and only where the value has the members it requires and the values it pins
  (`const` or a one-value `enum`), which decoding a struct alone does not check,
  strict or lenient. A member of an `enum` of several values is checked too:
  strict, a value outside it rules the alternative out; leniently, only while
  another alternative fits, as an API may add a value. An alternative that is a
  tagged union requires what every one of its own alternatives does, its tag
  only if each requires it.
- **Tagged unions** — a union whose alternatives differ only in a tag, a member
  each fixes to a string of its own, and in members of their own, such as
  Notion's blocks (`{"type": "paragraph", "paragraph": {...}}`), becomes one
  struct instead: the members all alternatives share, the tag, and one optional
  field per member of an alternative's own. An alternative's own members are at
  most one, named after its value, or more beside that one, such as a relation's
  `has_more`; one that several alternatives have must be the same in each. The
  tag is an enum type of its values, named after the struct and the tag (such as
  `BlockType` with `BlockTypeChildDatabase`), unless a part of an `allOf` types it
  already or a name is taken. The methods check that only members of the
  alternative the tag names are set, and that those it requires are there, as
  `null` if need be: decoding checks the members the object holds, encoding the
  fields set, where a member that may be `null` cannot be told from one left out. Encoding with the tag left empty sends the value whose member is
  set, and where an alternative may leave the tag out, decoding infers it so. Alternatives that are unions tagged alike count by their own alternatives,
  and as part of an `allOf` the union's fields join the struct's. A tagged struct
  that is an alternative of another union, there or in an `allOf`, is chosen by
  the members present, as an object is. Alternatives nothing else refers to get
  no type of their own.
- **allOf** — each part referenced by this schema alone is folded into its fields;
  a part other schemas share stays an embedded type. Properties the schema
  declares beside its `allOf` are fields too, after the parts'. A union among the parts is a
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
  learn from, then decodes it again without rejecting members the specification
  does not know. Only if that fails too does the call fail, so a response that
  merely holds more than the specification says still reaches the caller.
  Nothing the specification leaves open decodes into `any` then: the empty
  schema, an array without `items`, a free-form object and `not` alone become
  `struct{}`, so any value in them is recorded, and any but an object fails.
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
- **Read-only and write-only** — a property only responses carry (`readOnly`)
  or only requests (`writeOnly`) is never required, so a value that leaves it
  out still decodes, and a union's alternative is chosen without it. The client
  leaves read-only members out of a request body, and the server leaves
  write-only members out of a response, as their fields hold them or not, through
  unions and embedded parts alike; any other encoding keeps them.
- **Null** — a schema that is only ever `null` is `*struct{}`, and "X or null"
  is `X`, a pointer to `X` only by the rule for fields above, so that null and
  the zero value can differ, or where `X` decodes itself, such as a union or a
  tuple, which would refuse null.
- **Fixed parameters** — a required parameter that can take only one value, its
  `const` or the only value of its `enum`, is sent by the client itself: in the
  path, the query or the headers, such as Notion's `Notion-Version`. The caller
  never passes it. An `example` alone does not fix a value.
- **Query parameters** — every style OpenAPI defines, written by the client,
  read by the server and api.js alike. An array or an object of single values
  (a struct of its properties, or a map) is sent as one value joined by commas
  (`form`, not exploded), spaces (`spaceDelimited`) or pipes (`pipeDelimited`),
  an object as its names and values in turn; exploded, an array as one value
  per element, an object as one parameter per member, under the member's name,
  or with `deepObject` as `name[member]`. An exploded map holds every parameter
  no other does. `allowReserved` leaves reserved characters as they are, but for
  `&`, `#` and `+`, which would change how the query reads. A delimiter within a
  value is escaped, so every value arrives as it was sent. "X or an array of X"
  is sent as the array, any other union as a string, and `null` is dropped,
  since a query string cannot carry it. A duration is sent in whole seconds and
  an integer date-time as a Unix time, as in a body.
- **Authentication** — each operation sends the credential its own `security`
  names, else the document's: a bearer token or basic auth, read from the
  environment. Where a document uses both, `NewClient` requires at least one, and
  each operation fails before sending if its own is missing. An operation whose
  credentials are optional (`{}` beside a scheme) sends them all the same, and its
  replay test accepts a recording made without them.
- **Replay tests** — each recorded call is replayed against the generated client,
  which must send the recorded method, URL, body and headers. What identifies
  the caller is configuration, not part of the call: the `User-Agent`, the
  credential after `Authorization`'s scheme, and the headers the client sends
  from its options, such as Habitica's `X-Client` and an API key, are checked
  only to be there, so a call recorded by another app still matches.
- **Success responses** — an operation returns its success body as `*T`, or as
  `T` where `T` is already nilable: a slice, a map, or a named type of either. An
  operation whose success body is an empty object returns just `error`, and the
  server writes `{}`. Its body is read only in debug mode, where it is decoded so
  that anything in it fails loudly and is recorded. `XWithResult[R]` decodes
  into a type of the caller's own instead, leniently, since it declares only what
  it needs; the operation's own type is decoded strictly.
- **Binary responses** — a success body that is not text, such as a zip, a PDF,
  an image or a video, is returned as an `io.ReadCloser` that reads it as it
  arrives, never held whole in memory; the caller closes it. A text body that is
  not JSON is returned as `[]byte`. The server copies a returned reader into
  the response, and the JavaScript client returns such a body as a `Blob`, or
  text as a string.
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
