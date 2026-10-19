# go-auth

A small, self-hosted authentication service written in Go.

`go-auth` runs as a standalone HTTP service with your own PostgreSQL database.
Your application talks to it over JSON, so it can be written in any language.
It issues short-lived JWT access tokens signed with Ed25519, and long-lived
refresh tokens that rotate on every use and can be revoked. It publishes its
public keys as a JWKS document, so your services verify access tokens
themselves, without calling `go-auth` on every request, or ask `go-auth` to
check each one.

```
POST /v1/auth/register         create a user
POST /v1/auth/login            exchange credentials for tokens
POST /v1/auth/change-password  replace a password and revoke every session
POST /v1/auth/refresh          exchange a refresh token for a new token pair
POST /v1/auth/verify           check an access token on your service's behalf
POST /v1/auth/me               return the user an access token was issued to
GET  /.well-known/jwks.json    public keys for verifying access tokens
```

## What it does

- Registration and login with a username and password, and password changes
  that sign the user out of every session.
- Passwords stored as bcrypt hashes; refresh tokens stored only as a SHA-256
  hash.
- Single-use refresh tokens: presenting one twice revokes its whole session.
- Login and change password answer every credential failure the same way.
- Optional rate limiting per client address, structured logs with a request
  ID, and migrations embedded in the binary.

It has no email, OAuth or passwordless login, no password reset, no logout, no
roles or scopes, and no TLS of its own: run it behind a reverse proxy. Read
[Limitations](docs/system.md#limitations) before deciding whether it fits your
project.

## Quick start

Requires Docker.

```bash
git clone https://github.com/rafaelblt/go-auth.git
cd go-auth
docker compose -f docker-compose.dev.yml up --build
```

This starts PostgreSQL and the API, with the migrations applied, on
`http://localhost:8080`:

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"correct-horse"}'
```

## Documentation

Start with [How go-auth works](docs/system.md), then
[Getting started](docs/getting-started.md) and the
[API reference](docs/api/reference.md); [docs/README.md](docs/README.md) lists
every document. To contribute, read [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).
