# Contributing

`go-auth` is a small service with documentation that carries most of its
reasoning. This file is a router: it says what to read and what to run, and the
detail lives in `docs/`.

## Before writing code

Read **[Code conventions](docs/architecture/conventions.md)**. The patterns there
are specific enough to get wrong by guessing: two constructors per value object
(`New` for input, `Restore` for storage), `Value()` rather than `String()` on
anything holding a secret, `validation.Issues` from an input constructor but
`error` from a value generated internally, params structs past two arguments.

Then, for the layer you are touching, the document for it:
[Architecture overview](docs/architecture/overview.md) has the map.

## Rules that are not obvious from one file

- **Dependencies point inward.** `usecase` imports `port` and the domain
  packages, never `api`, `infra`, `bootstrap` or `config`.
- **Only `infra.SystemClock` calls `time.Now`.** Everything else takes a
  `port.Clock`, which is how expiry and reuse detection are tested without
  sleeping.
- **Only `internal/api`, `internal/bootstrap` and `cmd` log.** The domain, the
  use cases and the adapters return errors; the HTTP boundary logs them once,
  where the request ID and the outcome are both known. See
  [Logging](docs/architecture/logging.md#where-logging-happens).
- **Nothing logs a credential, a token or a password.** This holds because of the
  call sites, not a filter.
- **`docs/` describes only what exists.** An idea, a planned feature or an open
  question does not go there.

## Running the tests

```bash
go test ./...          # everything; needs a running Docker daemon
go test -race ./...    # with the race detector
```

Only the fast layers, with no database:

```bash
go test $(go list ./... | grep -vE 'infra/postgres|infra/migrate|testutil|tests/e2e')
```

`-short` skips nothing: no test checks `testing.Short()`. The full details, the
five test layers and the helper packages are in
[Testing](docs/development/testing.md).

## Before committing

```bash
gofmt -l .        # must print nothing
gofmt -w .        # format
go vet ./...
```

## Commits

The format, the ten allowed types and the scope names are in
[Commits](docs/development/commits.md). The short version: `type(scope): summary`,
a lowercase imperative summary under 72 characters, and a body that says *why*
whenever the subject leaves that unanswered — the diff already shows the *what*.

One change per commit, complete with its tests and the documentation it would
otherwise make wrong.

## Decision records

When the reasoning behind a change outgrows a commit body — it spans packages, or
there were real alternatives to weigh — it goes in a
[decision record](docs/development/decisions/README.md). That document says when a
record is warranted and when the reason belongs somewhere else instead. Most
changes apply an existing decision rather than make a new one.
