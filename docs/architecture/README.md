# Architecture

How `go-auth` is built, from the outside in. Start with the
[overview](overview.md) for the map, then follow whichever layer you need.

## The map

- **[Overview](overview.md)**: what kind of program this is, the layers, the
  dependency rule, the package map, and a request end to end.

## The layers

- **[Domain](domain/README.md)**: entities, value objects, and the rules each
  one enforces.
- **[Use cases](usecases/README.md)**: the ports they depend on, the shape
  they share, and register, login, change password, refresh and verify step by
  step.
- **[HTTP](http.md)**: routing, JSON bodies, middleware, and how errors become
  responses.
- **[Persistence](persistence/README.md)**: schema, migrations, repositories,
  transactions.
- **[Startup](startup.md)**: the composition root and the background tasks.

## Across the layers

- **[Tokens](tokens.md)**: the two token types, the signing keys, and the
  security properties they add up to.
- **[Logging](logging.md)**: every line the service writes, what its fields
  mean, and how to trace one request.
- **[Code conventions](conventions.md)**: the patterns repeated throughout the
  code, and why.

Code paths are relative to the repository root.
