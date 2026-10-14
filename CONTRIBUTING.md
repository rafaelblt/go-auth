# Contributing

## Before writing code

Read the [code conventions](docs/architecture/conventions.md): the patterns are
specific enough that guessing produces code that does not fit. Each package's
doc comment names the document that covers it, and the
[package map](docs/architecture/overview.md#packages) lists them all.

## Compatibility

v1 is a published contract, and must stay backward compatible. Client
applications and the services that verify tokens depend on:

- the `/v1` endpoints: their paths, status codes, error codes, and request and
  response fields ([API reference](docs/api/reference.md),
  [openapi.yaml](docs/api/openapi.yaml));
- the access token's header and claims, and the JWKS;
- the configuration variables and what they mean.

Adding is compatible: a new endpoint, a new response field, a new error code, a
setting with a default. Removing, renaming, or changing the type or meaning of
any of these is not. Neither is tightening a username or password rule: a
stored username is rebuilt with `NewUsername` when it is read, and login checks
credentials against the registration rules, so existing users would be locked
out. Data already stored must keep loading, and a released migration is never
edited: a schema change is a new migration.

## Documentation

- **The code is the source of truth.** The documents say what it does and why;
  when one disagrees with the code, fix the document, or the code if the code
  is what is wrong.
- **Every fact has one home.** A document that needs a fact explained elsewhere
  links to it rather than restating it.
- **A decision is explained where its code is documented**, in that package's
  section: what has to hold, why, and the alternatives that were really
  weighed, when there were any. One that spans packages lives with the code
  that enforces it, and the others link there.
- **A change updates the documents it makes wrong**, in the same commit.
- **`docs/` describes what exists**, not plans or ideas. A deprecation is
  documented, with what replaces it.

## Before committing

```bash
gofmt -l .      # must print nothing
go vet ./...
go test ./...   # needs Docker; see Testing for the subset that does not
```

[Testing](docs/development/testing.md) covers the test layers and helpers, and
[Commits](docs/development/commits.md) the message format and what goes in one
commit.
