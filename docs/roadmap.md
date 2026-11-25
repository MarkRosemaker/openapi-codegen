# Roadmap

Work not yet done. An entry is deleted once it is.

## No pointers in types only ever decoded

A field is a pointer where its zero value must be told apart from leaving it
out, so that a client can send `false`, `0` or `{}`. A type that only responses
use is never sent, and decoding rarely needs to tell an absent field from its
zero value, so its fields could drop the pointer. That needs to know which types
a request body can reach; a type both use keeps its pointers.

## Constructors that fill in fixed fields

A request parameter that can take only one value is sent by the client
itself. A request body's field that can take only one value, its `const` or
the only value of its `enum`, is still the caller's to set, such as a
discriminator of a union that is not a tagged struct, which infers its own.
A constructor, or a method that fills in
every such field left at its zero value, would spare the caller that, for a
struct used in a request body, or in one. Filling them in inside the
operation method instead would override what the caller set.

`MinimalLiteral` would need the same: it sets a union's first variant but not
the fields that tell the variants apart, so its value of a type that holds a
union may not decode. The generated tests of strict and lenient decoding leave
out every response type that holds one until it does.

## Path and header parameter styles

Query parameters are written in every style OpenAPI defines. A path parameter
is still a single value, with no `label` or `matrix` style and no array or
object, and a header parameter likewise; the server does not read header
parameters at all, so a service sees them empty.

## Unix times in milliseconds

An integer `date-time` is read as Unix seconds. Some APIs write milliseconds:
go-api-libs/finnhub's tick, NBBO and bid-ask timestamps ("List of timestamp in
UNIX ms.") and the websocket's trade times, and go-api-libs/yahoo-finance's
`dateEpochMs` and `firstTradeDateMilliseconds`. They are integers until
recordings show their values, since a spec's types rest on recordings: then a
custom format, `date-time-ms`, like `date-basic`, makes them a `time.Time` in
milliseconds, written and read through jsonutil helpers beside
`TimeMarshalIntUnix`, in a body and a query parameter alike. openapi-enrich
must keep the format when a recording's integer meets it.

## More of oapi-codegen's extensions

`x-go-name` and `x-enum-varnames` / `x-enumNames` work as in
[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/blob/main/docs/extensions.md),
a property's own extensions beside a `$ref` winning over the referenced
schema's. These would fit too, each as oapi-codegen has it, once a spec needs
it:

- **`x-go-type` / `x-go-type-import`:** a field's or a type's Go type, given,
  such as a decimal type for a number written as a string.
- **`x-go-type-skip-optional-pointer`:** an optional field without a pointer,
  where its zero value needs no telling apart from its absence; see also "No
  pointers in types only ever decoded".
- **`x-omitzero`:** `omitzero` on a field that would not otherwise have it.
- **`x-oapi-codegen-extra-tags`:** more struct tags on a field.
- **`x-deprecated-reason`:** a `Deprecated:` paragraph in a type's or field's
  doc comment.
- **`x-order`:** a field's place in its struct, where the spec's order of
  properties is not the one wanted.

The others do not fit: `x-omitempty` is `x-omitzero`'s older form,
`x-go-type-name` a second way to name a type beside `x-go-name`,
`x-go-json-ignore` a property better left out of the spec,
`x-oapi-codegen-only-honour-go-name` an unexported field, and
`x-oapi-codegen-enum-merge` a choice between ways of merging one property's
enums across an `allOf`'s parts, which the generator does not merge.
