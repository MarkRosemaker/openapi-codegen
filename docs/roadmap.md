# Roadmap

Work not yet done. An entry is deleted once it is.

## Exported names for lowercase components

A component named `error_api_400` or `idResponse` becomes an unexported Go
type, which a caller outside the package cannot name, for instance in
`errors.As`. Exporting them renames several hundred types in the Notion test
data and collides three times in each of notion-official and
notion-undocumented (`Date` and `date`, `TemplateMention` and
`templateMention`, `IconPageIcon` and `iconPageIcon`), so it needs a rule for
collisions, and it is the owner's call.

## No pointers in types only ever decoded

A field is a pointer where its zero value must be told apart from leaving it
out, so that a client can send `false`, `0` or `{}`. A type that only responses
use is never sent, and decoding rarely needs to tell an absent field from its
zero value, so its fields could drop the pointer. That needs to know which types
a request body can reach; a type both use keeps its pointers.
