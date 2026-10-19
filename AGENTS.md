# AGENTS.md

`go-auth` is a self-hosted authentication service: a standalone HTTP/JSON API
over its own PostgreSQL database, issuing Ed25519 JWT access tokens and
rotating, revocable refresh tokens. It is at v1.

This repository keeps its reasoning in `docs/` rather than in comments, so check
there before inferring intent from the code. Read only what the task needs:

1. **What the service does**: [How go-auth works](docs/system.md), the section
   of the process you are changing.
2. **How the code does it**: the [architecture overview](docs/architecture/overview.md)
   once, then the guide section of each package you touch, which its doc
   comment names. A section lists the package's files, so you need not open
   them all.
3. **A common change**, such as an endpoint, a setting or a migration: its
   [recipe](docs/architecture/recipes.md) lists every file it touches.
4. **Before writing code**: the [code conventions](docs/architecture/conventions.md).
5. **How to work here**: [CONTRIBUTING.md](CONTRIBUTING.md): the compatibility
   rule, the documentation rules, and what to run before committing.

Rules for agents:

- When a change cannot stay within the
  [v1 contract](CONTRIBUTING.md#compatibility), stop and ask rather than work
  around it.
- When a document disagrees with the code and it is the code that looks wrong,
  say so rather than change either silently.
- A change updates the documents it makes wrong, in the same commit; how to
  find them is in [CONTRIBUTING.md](CONTRIBUTING.md#documentation).
- Only commit or push when asked.
