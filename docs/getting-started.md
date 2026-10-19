# Getting started

Run the service, then walk through a register → login → refresh cycle.

## Requirements

- **Go** 1.26 or newer, to build from source.
- **PostgreSQL.** The tests run against 16, and the development stack uses 17.
- **Docker**, for the development stack and the tests that need a database.

## Option A: Docker Compose

`docker-compose.dev.yml` starts PostgreSQL and the API, and applies the
migrations on startup:

```bash
docker compose -f docker-compose.dev.yml up --build
```

The API listens on `http://localhost:8080`, and PostgreSQL on `localhost:5432`,
with the credentials in the compose file; data is kept in the `go-auth_pg_data`
volume. This stack is for development only: the credentials are hard-coded, and
its image carries the whole Go toolchain. To deploy, build the
[production image](#the-production-image).

## Option B: from source

Start a PostgreSQL database however you like, then:

```bash
export DATABASE_URL='postgres://user:password@localhost:5432/dbname?sslmode=disable'
export ADDRESS='localhost:8080'
export AUTO_MIGRATE='true'
export LOG_FORMAT='text'   # easier to read in a terminal than JSON

go run ./cmd
```

Every variable is in [Configuration](configuration.md).

## Migrations

The migrations are embedded in the binary, so the service can apply them
itself. With `AUTO_MIGRATE=true` it applies the pending ones at startup;
without it, it only checks. Either way, it refuses to start unless the schema
version is exactly the latest migration it carries:

```
get current schema version failed: schema not initialized: schema_migrations table not found
```

```
the database schema version (5) is behind the latest migration (6): apply the pending migrations
```

```
the database schema version (7) is ahead of the latest migration (6): the binary is older than the database
```

`AUTO_MIGRATE=true` is the right setting for development, and workable for a
single instance. To make a schema change a deliberate deploy step instead,
apply the migrations with the [golang-migrate](https://github.com/golang-migrate/migrate)
CLI, from the `migrations/` directory of the same commit as the binary, and
start the service with `AUTO_MIGRATE` off:

```bash
migrate -path ./migrations -database "$DATABASE_URL" up
```

The files and the version table are the ones the service uses, so what the
CLI applies is what the service checks for.

## First requests

### 1. Register a user

```bash
curl -sX POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"correct-horse"}'
```

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
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

The response has a new access token *and* a new refresh token. The old one is
now spent, and sending it again revokes the whole session
([client obligations](api/reference.md#client-obligations)).

### 4. Verify an access token in your own service

Fetch the public keys, cache them, and validate tokens in your service
([how, with examples](api/token-verification.md)):

```bash
curl -s http://localhost:8080/.well-known/jwks.json
```

Or let `go-auth` check each token, at the cost of a call on every request:

```bash
curl -sX POST http://localhost:8080/v1/auth/verify \
  -H 'Content-Type: application/json' \
  -d '{"access_token":"eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..."}'
```

The answer names the user the token was issued to, as `user_id`, or is
`401 invalid_token`.

### 5. Find out whose token it is

An access token carries the user ID and nothing else. To show the username,
ask for the user the token was issued to:

```bash
curl -sX POST http://localhost:8080/v1/auth/me \
  -H 'Content-Type: application/json' \
  -d '{"access_token":"eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..."}'
```

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

### 6. Change the password

```bash
curl -sX POST http://localhost:8080/v1/auth/change-password \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","current_password":"correct-horse","new_password":"battery-staple"}'
```

The old password stops working, and so does every refresh token issued before
the change: log in again with the new password.

## Deploying

Read [Limitations](limitations.md) first. The two things most likely to matter:

- `go-auth` has no TLS of its own, and its rate limiting is off unless you set
  [`RATE_LIMIT`](configuration.md#rate_limit): run it
  [behind a reverse proxy](limitations.md#run-it-behind-a-reverse-proxy).
- Set [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key),
  or the signing keys are stored unencrypted in the database
  ([why it matters](limitations.md#signing-keys-are-stored-in-the-database)).

### The production image

`Dockerfile` builds the image to deploy:

```bash
docker build -t go-auth .
```

Its runtime stage is `scratch`: the static binary, the CA certificates and a
`passwd` entry, with no shell, no package manager and no Go toolchain. The
migrations are embedded in the binary. Configure it through the environment:

```bash
docker run --rm \
  -e DATABASE_URL='postgres://user:password@host:5432/dbname?sslmode=require' \
  -e ADDRESS='0.0.0.0:8080' \
  -p 8080:8080 \
  go-auth
```

Inside a container, `ADDRESS` must bind `0.0.0.0`. The process runs as
`nobody`, so a port above 1024 is the safe choice: whether it may bind a lower
one depends on the container runtime. `EXPOSE 8080` is documentation; the port
served is the one in `ADDRESS`. `AUTO_MIGRATE` is off by default, so the
container refuses to start until the schema matches: set it, or
[apply the migrations as a deploy step](#migrations).

There is no health or readiness endpoint. `running app...` on standard output
is the readiness signal there is; for a liveness probe, a TCP check against
`ADDRESS` is the closest thing, and it will not notice a database that has gone
away.
