# Documentation

Every document, by what you are trying to do. Code paths are relative to the
repository root.

## Understanding and calling the service

Each topic holds its behaviour, its endpoints, the reasons behind them and its
limitations.

- **[How go-auth works](system.md)**: what the service is, its concepts, every
  endpoint and the topic that describes it, what it does not do, and the data
  it keeps.
- **[Accounts](accounts.md)**: users and passwords, and the endpoints that take
  a password.
- **[Sessions](sessions.md)**: what happens to a session after login, until
  the cleanup deletes it.
- **[Access tokens](access-tokens.md)**: checking them, yourself or through
  `go-auth`, and the keys that sign them.
- **[API conventions](api/conventions.md)**: what every endpoint shares, the
  errors included.
- **[OpenAPI spec](api/openapi.yaml)**: the same contract, machine-readable.

## Running the service

- **[Getting started](getting-started.md)**: run it, apply the migrations, make
  the first requests, deploy it.
- **[Configuration](configuration.md)**: every environment variable.
- **[Logging](logging.md)**: every line the service writes, and how to read
  them.

## Working on the code

- **[Architecture overview](architecture/overview.md)**: the layers, how to read
  the guides, and the package map, which names each package's guide.
- **[Code conventions](architecture/conventions.md)**: the patterns repeated
  throughout, and why.
- **[Recipes](architecture/recipes.md)**: the common changes, with every file
  they touch.
- One guide per layer, with a section per package: its files, the rules its
  code keeps, what changes with it, and the decisions behind it:
  [Domain](architecture/domain.md), [Use cases](architecture/usecases.md),
  [HTTP](architecture/http.md), [Persistence](architecture/persistence.md),
  [Tokens](architecture/tokens.md), [Startup](architecture/startup.md).

## Contributing

- **[CONTRIBUTING.md](../CONTRIBUTING.md)**: compatibility, documentation
  rules, what to run before committing.
- **[Testing](development/testing.md)** and **[Commits](development/commits.md)**.
