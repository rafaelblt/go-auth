# Documentation

Every document, by what you are trying to do. Code paths are relative to the
repository root.

## Understanding the service

- **[How go-auth works](system.md)**: what it does in each process, why, the
  data it keeps, and its limitations. The one place each fact about its
  behaviour is stated; the other documents link to it.

## Running the service

- **[Getting started](getting-started.md)**: run it, apply the migrations, make
  the first requests, build the production image.
- **[Configuration](configuration.md)**: every environment variable.
- **[Logging](logging.md)**: every line the service writes, and how to read
  them.

## Calling the API

- **[API reference](api/reference.md)**: endpoints, bodies, conventions.
- **[Error model](api/errors.md)**: the error shape and every code.
- **[OpenAPI spec](api/openapi.yaml)**: the same contract, machine-readable.
- **[Verifying access tokens](api/token-verification.md)**: with the JWKS, or by
  asking `go-auth`.

## Working on the code

- **[Architecture overview](architecture/overview.md)**: the layers, how to read
  the guides, and the package map, which names each package's guide.
- **[Code conventions](architecture/conventions.md)**: the patterns repeated
  throughout, and why.
- **[Recipes](architecture/recipes.md)**: the common changes, with every file
  they touch.
- One guide per layer, with a section per package: its files, the rules its
  code keeps, what changes with it, and the decisions behind it:
  [Domain](architecture/domain.md), [Use cases](architecture/usecases.md),
  [HTTP](architecture/http.md), [Persistence](architecture/persistence.md),
  [Tokens](architecture/tokens.md), [Startup](architecture/startup.md).

## Contributing

- **[CONTRIBUTING.md](../CONTRIBUTING.md)**: compatibility, documentation
  rules, what to run before committing.
- **[Testing](development/testing.md)** and **[Commits](development/commits.md)**.
