# go-auth

A small, self-hosted authentication service written in Go.

`go-auth` runs as a standalone HTTP service with your own PostgreSQL database.
Your application talks to it over JSON, so it can be written in any language.
It issues short-lived JWT access tokens signed with Ed25519, and long-lived
refresh tokens that rotate on every use and can be revoked. It publishes its
public keys as a JWKS document, so your services verify access tokens
themselves, without calling `go-auth` on every request, or, if they would
rather not, ask `go-auth` to check each one.

```
POST /v1/auth/register       create a user
POST /v1/auth/login          exchange credentials for tokens
POST /v1/auth/refresh        exchange a refresh token for a new token pair
POST /v1/auth/verify         check an access token on your service's behalf
GET  /.well-known/jwks.json  public keys for verifying access tokens
```

## What it does

- Registration and login with a username and password.
- Passwords stored as bcrypt hashes.
- Stateless access tokens: JWTs signed with EdDSA (Ed25519), valid for a short
  time.
- Stateful refresh tokens: single use, replaced on every refresh, stored only
  as a SHA-256 hash.
- Reuse detection: presenting a refresh token that was already used revokes
  its whole session.
- Access token checks on request, for services that would rather not verify
  tokens themselves.
- Login answers every failure the same way: one error for every cause, and a
  bcrypt comparison even when the account does not exist.
- Optional rate limiting of register, login and refresh per client address,
  at three levels.
- Structured logs, with an ID for each request.
- Migrations embedded in the binary, and a startup check that the database
  schema matches the binary.

## What it does not do

It has no email, OAuth or passwordless login, no password reset, no logout,
no roles or scopes, and no multi-tenancy. It has no TLS of its own, and its
rate limiting is off by default, so run it behind a reverse proxy. The signing keys are stored
in its database, unencrypted unless you set an encryption key. Read
[Limitations](docs/limitations.md) before deciding whether it fits your
project.

## Quick start

Requires Docker.

```bash
git clone https://github.com/rafaelblt/go-auth.git
cd go-auth
docker compose -f docker-compose.dev.yml up --build
```

This starts PostgreSQL and the API with the migrations applied, and the API
listens on `http://localhost:8080`.

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"correct-horse"}'
```

Full walkthrough: [Getting started](docs/getting-started.md).

## Documentation

| | |
|---|---|
| [Getting started](docs/getting-started.md) | Run the service and make your first requests |
| [Configuration](docs/configuration.md) | Every environment variable |
| [Limitations](docs/limitations.md) | What it does not do, and the trade-offs to know before deploying |
| [API reference](docs/api/reference.md) | Endpoints, request and response bodies |
| [Error model](docs/api/errors.md) | Status codes and error codes |
| [OpenAPI spec](docs/api/openapi.yaml) | The API contract in OpenAPI 3.1 |
| [Verifying access tokens](docs/api/token-verification.md) | Validating a JWT in your own service with the JWKS, or letting `go-auth` do it |
| [Architecture](docs/architecture/overview.md) | Layers, packages, dependency rules |
| [Code conventions](docs/architecture/conventions.md) | The patterns used throughout the code, and why |
| [Testing](docs/development/testing.md) | The test layers, how to run them, the helpers |
| [All documentation](docs/README.md) | Everything else |

## Contributing

Read [Code conventions](docs/architecture/conventions.md) before writing code,
and [CONTRIBUTING.md](CONTRIBUTING.md) for how to run the tests and shape a
commit.

## License

MIT. See [LICENSE](LICENSE).
