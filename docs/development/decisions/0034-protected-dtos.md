# A mapper is the only way to build a DTO

- **Status:** Accepted

## Invariant

Every use case lives in a *subpackage* of `internal/usecase`, and the DTOs and
their mappers in `internal/usecase` itself. That package boundary is the
guarantee: moving a use case up into `internal/usecase` would leave the
unexported fields in place and the guarantee gone. `tests/docs` checks it.

## Context

The DTOs in `internal/usecase/dtos.go` used to have exported fields, on the
argument that a DTO is built at the end of a use case, read once by the HTTP
layer and dropped, so immutability protects nothing. That holds for
immutability but misses construction: login and refresh built the access token
DTO by hand instead of calling its mapper, so the mapper's checks never ran,
and nothing noticed, because the DTOs happened to be right.

## Decision

DTO fields are unexported, with accessors
([Protected struct](../../architecture/conventions.md#protected-struct)). A use
case can write `usecase.UserDTO{}` but cannot fill a field, so every non-zero
DTO has gone through a mapper, and every mapper rejects a zero or nil entity.
`IsZero` then tells the HTTP layer whether the whole DTO was built, and it
panics when it maps a zero one.

Use case `Input` and `Output` keep exported fields: an `Output` is declared in
the same package as the use case that builds it, where unexported fields would
restrict nothing.

## Consequences

- Tests outside `internal/usecase` cannot write a DTO literal. HTTP tests get
  DTOs from `internal/testutil/apitest`, which builds an entity and calls the
  real mapper; an exported constructor for tests would reopen the gap.
- A mapper guarantees a DTO came from a valid entity, not from the right one:
  mapping the wrong user is still for the use case tests to catch.
