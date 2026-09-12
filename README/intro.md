Generated code has a reputation for being obviously generated. This module tries
hard not to earn it: output is run through `goimports` and `gofumpt`, names are
converted to Go conventions rather than transliterated, and the emitted types are
the ones you would have declared by hand.

Getting there depends on the specification being in good shape first, which is why
this module doesn't work from the raw document. It normalizes the spec through the
rest of the family before generating anything:

1. **Load** the specification, and any recorded HTTP interactions
2. **Validate** it
3. **[Flatten](https://github.com/MarkRosemaker/openapi-flatten)** — every meaningful type gets a name, so it can become a named Go type
4. **[Compress](https://github.com/MarkRosemaker/openapi-compress)** — duplicate schemas collapse, so the same shape doesn't become five Go types
5. **Build an intermediate representation**, resolving schemas to Go types
6. **Match recorded interactions** to operations, for round-trip tests
7. **Render and format** the output
