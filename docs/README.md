# Documentation

Every document, by what you are trying to do. Code paths are relative to the
repository root.

## Running the service

- **[Getting started](getting-started.md)**: run it, apply the migrations, make
  the first requests, build the production image.
- **[Configuration](configuration.md)**: every environment variable.
- **[Logging](logging.md)**: every line the service writes, and how to read
  them.
- **[Limitations](limitations.md)**: what it does not do, and the trade-offs to
  know before deploying it.

## Calling the API

- **[API reference](api/reference.md)**: endpoints, bodies, conventions.
- **[Error model](api/errors.md)**: the error shape and every code.
- **[OpenAPI spec](api/openapi.yaml)**: the same contract, machine-readable.
- **[Verifying access tokens](api/token-verification.md)**: with the JWKS, or by
  asking `go-auth`.

## Understanding the code

- **[Architecture overview](architecture/overview.md)**: the layers, the
  dependency rule, and the package map, which names each package's document.
- **[Code conventions](architecture/conventions.md)**: the patterns repeated
  throughout, and why.
- One document per layer, with the reasoning behind its design:
  [Domain](architecture/domain.md), [Use cases](architecture/usecases.md),
  [HTTP](architecture/http.md), [Persistence](architecture/persistence.md),
  [Tokens](architecture/tokens.md), [Startup](architecture/startup.md).

## Contributing

- **[CONTRIBUTING.md](../CONTRIBUTING.md)**: compatibility, documentation
  rules, what to run before committing.
- **[Testing](development/testing.md)** and **[Commits](development/commits.md)**.
