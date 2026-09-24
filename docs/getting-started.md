# Getting started

This guide gets the service running and walks through a full
register → login → refresh cycle.

## Requirements

|            |                                                              |
| ---------- | ------------------------------------------------------------ |
| Go         | 1.26 or newer (`go.mod` targets `go 1.26.0`)                 |
| PostgreSQL | 16 or newer                                                  |
| Docker     | for the development compose stack and the integration tests |

16 is the floor, and it is the version the tests run against: the testcontainer
is `postgres:16`. The development compose stack uses `postgres:17-alpine`, so
both are exercised, 16 by the test suite and 17 by everyday development.

## Option A: Docker Compose

`docker-compose.dev.yml` starts PostgreSQL and the API together, and applies
the migrations on startup.

```bash
docker compose -f docker-compose.dev.yml up --build
```

The API listens on `http://localhost:8080`, and PostgreSQL on
`localhost:5432`, with the credentials set in the compose file. Data is kept
in the `goauth_pg_data` volume between restarts.

This stack is for development only: the credentials are hard-coded, and
`Dockerfile.dev` builds an image that contains the whole Go toolchain. For
deployment, build the [production image](#the-production-image) from
`Dockerfile`.

## Option B: from source

Start a PostgreSQL database however you like, then:

```bash
export DATABASE_URL='postgres://user:password@localhost:5432/dbname?sslmode=disable'
export ADDRESS='localhost:8080'
export AUTO_MIGRATE='true'
export LOG_FORMAT='text'   # easier to read in a terminal than JSON

go run ./cmd
```

Every variable, and its default, is in [Configuration](configuration.md).

## Migrations

The SQL migrations are embedded in the binary, so the service can apply them
itself, with no migration tool installed and no `.sql` files on disk.

- With `AUTO_MIGRATE=true`, the service applies pending migrations at startup.
- Without it, the service only checks the schema.

Either way, the service refuses to start unless the database schema version
is exactly the latest migration in the binary:

```
get current schema version failed: schema not initialized: schema_migrations table not found
```

```
the database schema version (3) is behind the latest migration (4): apply the pending migrations
```

```
the database schema version (5) is ahead of the latest migration (4): the binary is older than the database
```

A binary that started against a schema it was not built for would fail one
request at a time, as `500` responses. Failing at startup instead makes a
deploy visibly not roll out.

### Applying migrations as a deploy step

`AUTO_MIGRATE=true` is the only way *this binary* applies migrations: there is
no migration subcommand and no flag. It is the right setting for development,
and workable for a single-instance deployment, where the process that migrates
is the only one there is.

To make a schema change a deliberate step instead, apply the migrations
yourself with the [golang-migrate](https://github.com/golang-migrate/migrate)
CLI, against the `migrations/` directory of this repository, and start the
service with `AUTO_MIGRATE` off:

```bash
migrate -path ./migrations -database "$DATABASE_URL" up
```

Needing the checkout as well as the binary is the cost of keeping migrations
out of startup. The files are the ones embedded in the binary, and the version
table is the same `schema_migrations`, so what the CLI applies is what the
service checks for. Keep the checkout and the binary on the same commit, or
the startup check will refuse the mismatch.

## First requests

### 1. Register a user

```bash
curl -sX POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"correct-horse"}'
```

```json
{
  "user": {
    "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37",
    "username": "alice",
    "status": "active",
    "created_at": "2026-09-08T12:00:00Z",
    "updated_at": "2026-09-08T12:00:00Z"
  }
}
```

Registering does **not** return tokens. Log in next.

### 2. Log in

```bash
curl -sX POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"correct-horse"}'
```

```json
{
  "access_token": {
    "value": "eyJhbGciOiJFZERTQSIsImtpZCI6Ii4uLiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-09-08T12:30:00Z",
    "expires_in": 1799
  },
  "refresh_token": {
    "value": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01",
    "expires_at": "2026-09-15T12:00:00Z",
    "expires_in": 604800
  }
}
```

Keep the refresh token somewhere private that survives restarts. Send the
access token to your own services.

### 3. Refresh

```bash
curl -sX POST http://localhost:8080/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01"}'
```

The response has a new access token *and* a new refresh token. The old
refresh token is now spent, and sending it again revokes the whole session.
See [Authentication flows](architecture/usecases/refresh.md).

### 4. Verify an access token in your own service

```bash
curl -s http://localhost:8080/.well-known/jwks.json
```

```json
{
  "keys": [
    {
      "kty": "OKP",
      "crv": "Ed25519",
      "x": "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
      "use": "sig",
      "alg": "EdDSA",
      "kid": "NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"
    }
  ]
}
```

Fetch this document, cache it, and validate tokens in your service. How to do
it, with examples: [Verifying access tokens](api/token-verification.md).

## Deploying

Before deploying, read [Limitations](limitations.md). The two things most
likely to matter:

- `go-auth` has no TLS, rate limiting or request size limit of its own. Run it
  [behind a reverse proxy](limitations.md#run-it-behind-a-reverse-proxy).
- The signing key is held in memory, so every restart invalidates all
  outstanding access tokens, and the service runs as
  [a single replica](limitations.md#one-replica-and-a-restart-invalidates-access-tokens).

### The production image

`Dockerfile`, not `Dockerfile.dev`, builds the image to deploy:

```bash
docker build -t go-auth .
```

It is multi-stage, and the runtime stage is `scratch`: the static binary, the CA
certificates, a `passwd` entry, and nothing else. Around 12 MB, with no shell, no
package manager and no Go toolchain. The migrations are embedded in the binary,
so the image carries its own schema and needs no `.sql` files.

Configure it entirely through the environment:

```bash
docker run --rm \
  -e DATABASE_URL='postgres://user:password@host:5432/dbname?sslmode=require' \
  -e ADDRESS='0.0.0.0:8080' \
  -p 8080:8080 \
  go-auth
```

`ADDRESS` must bind `0.0.0.0` inside a container, and a port above 1024, since
the process runs as `nobody`. `EXPOSE 8080` in the `Dockerfile` is documentation:
the port actually served is the one in `ADDRESS`.

`AUTO_MIGRATE` is off by default, so the container will refuse to start until the
schema matches the binary. Either set it, or
[apply the migrations as a deploy step](#applying-migrations-as-a-deploy-step).

There is no `/health` and no `/ready` endpoint
([why](limitations.md#out-of-scope)). `running app...` on standard output is the
readiness signal available; for a liveness probe, a TCP check against `ADDRESS` is
the closest thing, and it will not notice a database that has gone away.

## Where to next

- [Configuration](configuration.md): token lifetimes, bcrypt cost, log
  format.
- [API reference](api/reference.md): the full contract.
- [Architecture overview](architecture/overview.md): how the code is laid
  out.
