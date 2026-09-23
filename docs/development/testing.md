# Testing

The tests run at six layers, and two of them need Docker.

## Layers

| Layer | Where | Needs Docker | Speed |
|---|---|---|---|
| Domain | `internal/domain/*`, `internal/validation`, `internal/shared` | no | instant |
| Use case | `internal/usecase/*` | no | instant |
| HTTP | `internal/api` | no | instant |
| Infrastructure | `internal/infra/postgres`, `internal/infra/migrate`, `internal/testutil/...` | **yes** | slow |
| End to end | `tests/e2e` | **yes** | slow |
| Documentation | `tests/docs` | no | instant |

Domain, use case and HTTP tests use fakes throughout, and run in
milliseconds. Infrastructure and end-to-end tests start a real PostgreSQL in a
[testcontainer](https://golang.testcontainers.org/), so they need a running
Docker daemon. Without one, they fail with
`Docker is unavailable (is it running?)`.

## Running them

```bash
go test ./...                       # everything (needs Docker)
go test ./internal/domain/user/...  # one package
go test ./tests/e2e/...             # end to end only
go test -race ./...                 # with the race detector
go test -run TestLogin ./...        # by name
```

To run only the fast tests, leave out the packages that need a database:

```bash
go test $(go list ./... | grep -vE 'infra/postgres|infra/migrate|testutil|tests/e2e')
```

`-short` does not skip anything: no test checks `testing.Short()`.

## Documentation tests

`tests/docs` checks the documentation rules that are worth not having to
remember. There is one: every decision record whose status is `Accepted` must be
mentioned by some `.go` file, which is the rule the
[decisions README](decisions/README.md#writing-one) states. It reads `docs/` and
the `.go` files as text and does not compile anything, so it is instant and needs
no database.

Keep this package small. A rule earns a test here when forgetting it is easy and
the breakage is invisible; anything else belongs in a document.

## Writing a test

1. **Test at the lowest layer that can prove it.** A domain rule belongs in a
   domain test, not an end-to-end test.
2. **Use the existing helpers**, and extend them rather than building fakes in
   the test.
3. **Table-driven** for several similar cases.
4. **`require` to stop, `assert` to continue.** `require` for a precondition
   whose failure makes the rest meaningless, `assert` for independent checks.

## Helper packages

| Package | Provides |
|---|---|
| `internal/testutil` | Test database, HTTP client, JSON decoding |
| `internal/testutil/porttest` | Fakes for the ports the use cases depend on |
| `internal/testutil/usertest` | `User` factories and musts |
| `internal/testutil/passwordtest` | `Password` factories and musts |
| `internal/testutil/sessiontest` | `Session` and `RefreshToken` factories and musts |
| `internal/testutil/apitest` | DTO factories for HTTP tests |
| `internal/testutil/postgrestest` | Test pools, database reset, and inserts that bypass the repositories |
| `internal/testutil/migratetest` | Test database with the migrations applied |

### Fakes

One hand-written fake per port, in `porttest`. They are small, read as
ordinary Go, and the compiler catches it as soon as an interface changes, with
no generation step or mocking DSL to learn.

- **Readers** are in-memory stores: a test seeds them (`InsertUser`,
  `InsertPassword`, `Insert`) and the use case finds what was seeded.
- **Writers** record what they were given (`SavedUsers`, `Adds`,
  `MarkedUsed`, `Updates`), so a test checks what a use case persisted.
- **`SetError`** makes a fake fail, to test error paths.
- **`FakeClock.SetNow`** fixes the time, which is how expiry is tested without
  sleeping.

Because they record their calls, a test can check the interaction as well as
the result: that login checked the password against the dummy hash when the
user was missing, for instance (`FakePasswordChecker.Calls`).

### Factories and musts

```go
usr := usertest.NewUser(t, nil)                  // a valid user, with defaults
usr = usertest.NewUser(t, func(p *user.RestoreParams) {
    p.CreatedAt = createdAt                      // override only what matters
})
username := usertest.MustUsername(t, "alice")    // fails the test on error
```

A factory builds an entity from its restore params, and the optional
function edits those params before the entity is restored, so the result
still goes through validation. A test states only the field it is about.
Musts remove error handling for values that cannot fail, so a test reads as
the scenario it tests.

## Use case tests

Each use case package has a `helper_test.go` with a `TestHelper` holding every
fake, with defaults:

```go
helper := NewTestHelper(t)
usr, pwd := helper.GetUserAndPassword()   // seeds the fakes

output, err := helper.UseCase().Execute(ctx, login.Input{
    Username: usr.Username().String(),
    Password: pwd.Value(),
})
```

No database, no HTTP. A test changes only the fake it cares about.

## HTTP tests

`internal/api` tests cover each piece of the HTTP layer on its own: decoders,
encoders, error translation and the JSON writer. The use case adapter is
tested with a fake use case defined in its test file. No HTTP test runs a real
use case: the end-to-end tests cover that path.

DTO fields are unexported, so a test cannot write a DTO literal. The
factories in `apitest` build one through the real mapper:

```go
out := login.Output{
    AccessToken:  apitest.NewAccessTokenDTO(t),
    RefreshToken: apitest.NewRefreshTokenDTO(t),
}
usr := apitest.NewUserDTO(t, func(p *user.RestoreParams) { p.Username = usertest.MustUsername(t, "alice") })
```

Only a zero DTO, such as `usecase.UserDTO{}`, can be written directly, which
is what the tests of the zero-DTO panics need.

## Test databases

`internal/testutil/database.go` runs a `postgres:16` testcontainer.

```go
db, err := testutil.NewDatabase(ctx)          // you terminate it
db := testutil.NewDatabaseForTest(t, ctx)     // terminated by t.Cleanup
```

The container is ready once PostgreSQL has logged
"database system is ready to accept connections" **twice**. It logs it once
while initialising, and again when it actually accepts connections; waiting
for the first one caused flaky connection errors.

**The database starts empty, with no migrations.** The migration tests need a
database their migrations have not touched. For a migrated one, use
`migratetest.NewDatabaseWithMigrations(ctx)`.

**Starting a container takes seconds**, so each package starts one in
`TestMain` and resets it between tests:

- `internal/infra/postgres` uses `postgrestest.PoolFactory`. `Acquire` drops
  and recreates the schema before each test; `AcquireWithMigrations` then
  applies the migrations.
- `tests/e2e` empties every table except `schema_migrations` with
  `postgrestest.TruncateTables` before each test.

## End-to-end tests

`tests/e2e` runs the **real application**, with the real bootstrap, router and
repositories, against a migrated testcontainer, and drives it over HTTP.

`TestMain` builds and starts the app once. Each test calls
`testApp.NewEnv(t)`, which empties the tables so tests do not affect each
other.

```go
func TestSomething(t *testing.T) {
    env := testApp.NewEnv(t)
    resp := env.Client.Post(t, "v1/auth/login", body)
    // ...
}
```

The `TestEnv` holds:

- `Client`: an HTTP client for the running server;
- `Fixtures`: inserts data straight into the database, bypassing the API;
- `Asserts`: checks the database, to verify what an endpoint stored.

The suite uses `BcryptCost: 6`, so hashing does not dominate the login tests.

It covers register, login, refresh, JWKS and unknown paths, and in
`mix_test.go`, sequences across endpoints, including rotation and reuse
detection, the flow most worth testing end to end.
