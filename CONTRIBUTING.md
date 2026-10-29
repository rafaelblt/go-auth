# Contributing

## Before writing code

Read the [code conventions](docs/architecture/conventions.md): the patterns are
specific enough that guessing produces code that does not fit. Each package's
doc comment names the guide section that covers it, and the
[package map](docs/architecture/overview.md#packages) lists them all.

## Compatibility

v1 is a published contract, and must stay backward compatible. Client
applications and the services that verify tokens depend on:

- the `/v1` endpoints: their paths, status codes, error codes, and request and
  response fields ([Endpoints](docs/system.md#endpoints),
  [API conventions](docs/api/conventions.md),
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
- **Every fact has one home**, by its subject. Any other document that needs
  it links there, even when a reader would rather not click:

  | Fact | Home |
  |---|---|
  | What the service does about one subject, why, what a client sends and receives for it, and its limitations | the subject's topic document: [Accounts](docs/accounts.md), [Sessions](docs/sessions.md), [Access tokens](docs/access-tokens.md) |
  | What every endpoint shares: content type, request bodies, the error shape and codes, rate limiting | [API conventions](docs/api/conventions.md), mirrored in [openapi.yaml](docs/api/openapi.yaml) |
  | The concepts, the list of endpoints, what is out of scope, the data kept | [How go-auth works](docs/system.md) |
  | A setting, its values and its default | [Configuration](docs/configuration.md) |
  | A log line, its fields, and what each reason means | [Logging](docs/logging.md) |
  | Running and deploying the service | [Getting started](docs/getting-started.md) |
  | How the code does something: files, types, rules, decisions | the package's section in its [architecture guide](docs/architecture/overview.md#packages) |
  | The steps of a common change | [Recipes](docs/architecture/recipes.md) |

  The subject is what the fact is about, wherever it is needed: a username
  rule belongs to accounts, a refresh token to sessions, the content type to
  the API conventions. The topic documents name no code. The OpenAPI spec is
  the one deliberate copy, checked against the router.
- **Elsewhere, a fact is a link.** A clause may name it, but its details,
  numbers and reasons stay in its home. "The access tokens already issued keep
  working until they expire, for up to `ACCESS_TOKEN_TTL`" written outside
  [Access tokens](docs/access-tokens.md) is a copy; "it does not reach the
  access tokens already issued", with a link, is not.
- **An HTTP status appears only where the contract is described**: an
  endpoint's section, and the [API conventions](docs/api/conventions.md).
  Everywhere else, name the error code (`invalid_token`), or, in a guide, the
  error in the code (`ErrTokenInvalid`).
- **A reason is given only where a reader would otherwise be tempted to change
  the rule**: one that looks arbitrary, stricter than expected, or easy to
  "fix" by breaking something. One or two sentences, in a paragraph of its own
  at the end of the section, starting with **Why:**, never between the steps.
- **No lists or counts of things that grow.** "Register, login and refresh" or
  "seven endpoints" goes wrong with the next endpoint: link to the table that
  lists them.
- **No function bodies.** A guide names files, types and the rules the code
  keeps, not how a function is written, which the code says better.
- **A guide's *Rules the code keeps* hold what exists because of the code**:
  types, the order of calls, locks, sentinels, which function takes which
  step. The behaviour they implement is a link to its topic document.
- **A decision about the code is explained where its code is documented**, in
  that package's section: what has to hold, why, and the alternatives that
  were really weighed, when there were any. One that spans packages lives with
  the code that enforces it, and the others link there. The alternatives
  weighed go here, not in the topic documents.
- **A change updates the documents it makes wrong**, in the same commit. To
  find them: read the *When you change it* of each package section you touched,
  follow its recipe if there is one, and search `docs/`, the READMEs and the
  OpenAPI spec for the names the change affects, such as a log reason, an error
  code, a variable, a path or a field.
- **`docs/` describes what exists**, not plans or ideas. A deprecation is
  documented, with what replaces it.

## Before committing

```bash
gofmt -l .      # must print nothing
go vet ./...
go test ./...   # needs Docker; see Testing for the subset that does not
```

[CI](.github/workflows/ci.yml) runs the same checks, the tests with `-race`, on
every push to `main` and every pull request.

[Testing](docs/development/testing.md) covers the test layers and helpers, and
[Commits](docs/development/commits.md) the message format and what goes in one
commit.
