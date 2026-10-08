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

## Path and header parameter styles

Query parameters are written in every style OpenAPI defines. A path parameter
is still a single value, with no `label` or `matrix` style and no array or
object, and a header parameter likewise; the server does not read header
parameters at all, so a service sees them empty.
