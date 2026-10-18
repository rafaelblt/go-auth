# Testing

## Layers

| Layer | Where | Needs Docker |
|---|---|---|
| Domain | `internal/domain/*`, `internal/validation`, `internal/shared` | no |
| Use case | `internal/usecase/*` | no |
| HTTP | `internal/api` | no |
| Infrastructure | `internal/infra/postgres`, `internal/infra/migrate`, `internal/testutil/...` | **yes** |
| End to end | `tests/e2e` | **yes** |
| Documentation | `tests/docs` | no |

The layers without Docker use fakes and run in milliseconds. The others start a
real PostgreSQL in a [testcontainer](https://golang.testcontainers.org/), and
fail with `Docker is unavailable (is it running?)` without a Docker daemon.

## Running them

```bash
go test ./...          # everything; needs Docker
go test -race ./...    # with the race detector
go test $(go list ./... | grep -vE 'infra/postgres|infra/migrate|testutil|tests/e2e')   # only what needs no database
```

`-short` skips nothing: no test checks `testing.Short()`.

## Writing a test

1. **Test at the lowest layer that can prove it.** A domain rule belongs in a
   domain test, not an end-to-end test.
2. **Use the helpers** below, and extend them rather than build fakes in a test.
3. **Table-driven** for several similar cases.
4. **`require` to stop, `assert` to continue**: `require` for a precondition
   whose failure makes the rest meaningless.

## Helper packages

| Package | Provides |
|---|---|
| `internal/testutil` | The test database, an HTTP client, JSON decoding |
| `internal/testutil/porttest` | A fake for each port |
| `internal/testutil/usertest`, `passwordtest`, `sessiontest` | Entity factories and musts |
| `internal/testutil/apitest` | DTOs for HTTP tests |
| `internal/testutil/postgrestest` | Test pools, database reset, inserts that bypass the repositories |
| `internal/testutil/migratetest` | A test database with the migrations applied |

### Fakes

One hand-written fake per port, in `porttest`: small, ordinary Go, and the
compiler flags it as soon as an interface changes. Readers are in-memory stores
a test seeds; writers record what they were given, so a test checks what a use
case persisted and how (`FakePasswordChecker.Calls` shows that login checked the
dummy hash for a missing user); `SetError` makes a fake fail; and
`FakeClock.SetNow` fixes the time.

Each use case package has a `helper_test.go` with a `TestHelper` holding every
fake, with defaults, so a test changes only the fake it cares about.

### Factories and musts

```go
usr := usertest.NewUser(t, nil)                  // a valid user, with defaults
usr = usertest.NewUser(t, func(p *user.RestoreParams) {
    p.CreatedAt = createdAt                      // override only what matters
})
username := usertest.MustUsername(t, "alice")    // fails the test on error
```

A factory edits the restore params and restores the entity, so the result still
goes through the constructor's checks. A must removes the error handling of a
value that cannot fail, so a test reads as its scenario.

### HTTP tests

DTO fields are unexported, so a test cannot write a DTO literal
([DTOs](../architecture/usecases.md#dtos)). `apitest` builds DTOs through the
real mappers; an exported constructor for tests would let any code skip them.
The HTTP tests cover each piece of `internal/api` on its own, with a fake use
case for the adapter; the end-to-end tests cover the path through a real one.

## Test databases

`internal/testutil/database.go` runs a `postgres:16` container, ready once
PostgreSQL has logged "database system is ready to accept connections"
**twice**: it logs it once while initialising, and waiting for that one caused
flaky connection errors. The database starts empty; for a migrated one, use
`migratetest`.

Starting a container takes seconds, so each package starts one in `TestMain`
and resets it between tests:

- `internal/infra/postgres` uses `postgrestest.PoolFactory`: `Acquire` drops
  and recreates the schema, and `AcquireWithMigrations` then migrates it.
- `tests/e2e` empties every table but `schema_migrations` and `signing_keys`
  with `postgrestest.TruncateTables`. The signing keys stay because the running
  app holds them.

## End-to-end tests

`tests/e2e` runs the real application, with the real bootstrap, router and
repositories, against a migrated container, and drives it over HTTP. `TestMain`
starts it once, with the default configuration, so rate limiting is off, and
`BcryptCost` 6, so hashing does not dominate. Each test calls
`testApp.NewEnv(t)`, which empties the tables and gives it a `Client`,
`Fixtures` that insert straight into the database, and `Asserts` that check it.

Some tests start a second app: with `RATE_LIMIT` `strict` on `localhost:8081`,
since limits on the shared app would interfere with every other test; over the
same database, to check that two instances share the signing keys; and over
empty databases, to check that startup migrates before it reads the keys, and
that sealed seeds need the encryption key. `mix_test.go` runs sequences across
endpoints, rotation and reuse detection included.

## Documentation tests

`tests/docs` checks what holds the documents and the code together, reading
both as text. Each test guards a rule that is easy to forget and whose breakage
nothing else would report:

| Test | Rule |
|---|---|
| `TestMarkdownLinksResolve` | Every relative link between documents resolves, file and anchor |
| `TestDocLinksInCodeCommentsResolve` | Every `docs/….md` path in a comment exists, anchor included |
| `TestOpenAPISpecMatchesTheRoutes` | The routes in `docs/api/openapi.yaml` are the ones `internal/api/router.go` registers |
| `TestEnvExampleMatchesTheConfigVariables` | The variables in `.env.example` are the ones `internal/config/env_vars.go` declares, commented out or not |
| `TestEveryPackageHasAPackageComment` | Every package has a doc comment |
| `TestNoUseCaseLivesInTheUsecasePackageItself` | No use case sits in `internal/usecase` itself, where it could fill a DTO by hand ([DTOs](../architecture/usecases.md#dtos)) |

A rule the compiler, a unit test or a database constraint already enforces
gets no test here: it would only be a second place to update.
