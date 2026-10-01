# Roadmap

Work not yet done. An entry is deleted once it is.

## A union nested in the union of an allOf

`pagePropertyValueWithIdResponse`, in notion-official and notion-undocumented,
is an `allOf` whose union has a union among its alternatives
(`simpleOrArrayPropertyValueResponse`). Its methods return an "unimplemented"
error. Every leaf of the nested union fixes `type` to a value of its own, so the
leaves could be chosen by `type` like the other alternatives, setting the field of
the nested union as well.

## Exported names for lowercase components

A component named `error_api_400` or `idResponse` becomes an unexported Go
type, which a caller outside the package cannot name, for instance in
`errors.As`. Exporting them renames several hundred types in the Notion test
data and collides three times in each of notion-official and
notion-undocumented (`Date` and `date`, `TemplateMention` and
`templateMention`, `IconPageIcon` and `iconPageIcon`), so it needs a rule for
collisions, and it is the owner's call.

## Pointers for nullable fields

"X or null" is `*X` today for every X that is not already nilable, strings
included. Whether a nullable string should rather be a plain `string`, where
empty and null usually mean the same to a caller, is undecided. One proposal:
pointers for nullable numbers and booleans only.
