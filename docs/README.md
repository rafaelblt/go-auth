# Documentation

Documentation for `go-auth`, organised by what you are trying to do.

## Running the service

You want to deploy `go-auth` and point an application at it.

- **[Getting started](getting-started.md)**: run it with Docker Compose or
  from source, apply the migrations, make your first requests.
- **[Configuration](configuration.md)**: every environment variable, with its
  default, its format, and what goes wrong if it is set incorrectly.
- **[Limitations](limitations.md)**: what the service does not do, and the
  trade-offs to know before deploying it.

## Calling the API

You are writing a client for the HTTP API.

- **[API reference](api/reference.md)**: endpoints, request and response
  bodies, status codes.
- **[Error model](api/errors.md)**: the error shape, how errors map to
  status codes, and every error code.
- **[OpenAPI spec](api/openapi.yaml)**: the same contract, machine-readable
  (OpenAPI 3.1), for generating clients and trying requests.
- **[Verifying access tokens](api/token-verification.md)**: how to validate a
  JWT in your own service with the JWKS endpoint, with examples.

## Understanding the code

You are reading or changing the source. Each layer has its own document; the
[architecture index](architecture/README.md) lists them all.

- **[Architecture overview](architecture/overview.md)**: layers, dependency
  rules, package map, a request end to end.
- **[Domain](architecture/domain/README.md)**: entities, value objects, and
  the rules each one enforces.
- **[Use cases](architecture/usecases/README.md)**: the ports they depend on,
  and register, login and refresh step by step.
- **[Tokens](architecture/tokens.md)**: the two token types, the signing keys,
  and the security properties.
- **[HTTP](architecture/http.md)**: routing, JSON, middleware, error
  translation.
- **[Persistence](architecture/persistence/README.md)**: database schema,
  migrations, repositories, transactions.
- **[Startup](architecture/startup.md)**: the composition root and the
  background tasks.
- **[Logging](architecture/logging.md)**: every line the service writes, what
  its fields mean, and how to trace one request.
- **[Code conventions](architecture/conventions.md)**: the patterns used
  throughout the code, and why.

## Contributing

Start with **[CONTRIBUTING.md](../CONTRIBUTING.md)** in the repository root,
which routes to the documents below.

- **[Commits](development/commits.md)**: the message format, the types and
  scopes, and what belongs in one commit.
- **[Testing](development/testing.md)**: the test layers, how to run them, and
  the test helpers.
- **[Decision records](development/decisions/README.md)**: the few decisions
  whose reasoning is too detailed for these documents and cannot be read from
  the code.

Code paths are relative to the repository root.
