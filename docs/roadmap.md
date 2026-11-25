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

## Server routes for a parameter inside a segment

A path parameter can share its segment with fixed text, as go-api-libs/ibkr's
`/portfolio/U{u}/summary` and sec's `/api/xbrl/companyfacts/CIK{cik}.json` do.
The client builds such a path, but `RegisterService` hands the template to
`http.ServeMux` as it is, and a wildcard there must be the whole segment, so
registering the service panics. The server would route the whole segment and
strip the fixed text itself, answering 404 where it is missing.
