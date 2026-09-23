# DTOs are protected structs

- **Status:** Accepted
- **Date:** 2026-09-19
- **Areas:** api

## Context

`internal/usecase/dtos.go` had exported fields, the one exception to the
[protected-struct convention](../../architecture/conventions.md#protected-struct).
The exception was argued from lifetime: a DTO is built at the end of a use
case, read once by the HTTP layer and dropped, so keeping it immutable
protects nothing.

That argument holds for immutability but misses construction. With exported
fields, a use case could build a DTO by hand instead of calling its mapper,
and login and refresh did exactly that for the access token DTO. The DTOs were
right, so no test failed, but the mapper's empty-token check never ran. The
HTTP layer could not tell either: its `IsZero` checks look at one field, so a
DTO with an ID and nothing else passed as valid.

## Decision

DTO fields are unexported, with accessors named after them. A mapper in
`internal/usecase` is the only way to build a DTO that is not zero.

This works because of where the code lives. DTOs and their mappers are in
`internal/usecase`, and every use case is in a subpackage of it, so a use case
can write `usecase.UserDTO{}` but cannot fill a single field. Every non-zero
DTO has passed through a mapper, and every mapper rejects a zero or nil
entity. `IsZero` then tells the HTTP layer whether the whole DTO was built
correctly, not just one field.

Use case `Input` and `Output` keep exported fields. The protection only binds
code outside the declaring package, and an `Output` is declared in the same
package as the use case that builds it, so unexported fields would not stop
that use case from building it any way it wants. They would restrict only
the HTTP layer, which only reads it. An `Output` missing a DTO is still
caught: the HTTP layer panics when it maps a zero DTO.

## Alternatives considered

- **Keep exported fields** — how DTOs are built stays a matter of discipline,
  and the HTTP layer's zero check stays partial.
- **Also protect `Output`** — no guarantee in return, as above, unless outputs
  move to a package apart from their use cases, which costs more in
  organisation than a check one layer earlier is worth.

## Consequences

- Tests outside `internal/usecase` can no longer write a DTO literal. HTTP
  tests get DTOs from the factories in `internal/testutil/apitest`, which build
  a domain entity and call the real mapper. HTTP tests now depend on the domain
  test factories, and there is no shortcut around that: an exported
  constructor for tests would reopen the gap.
- A mapper guarantees a DTO came from a valid entity, not from the right one.
  Mapping the wrong user is still something only the use case tests catch.
- Revisit if a use case ever has to live in `internal/usecase` itself, or if
  DTOs move out of it: the guarantee depends on that package boundary.
