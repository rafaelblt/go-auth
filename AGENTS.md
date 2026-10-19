# AGENTS.md

`go-auth` is a self-hosted authentication service: a standalone HTTP/JSON API
over its own PostgreSQL database, issuing Ed25519 JWT access tokens and
rotating, revocable refresh tokens. It is at v1: register, login, change
password, refresh, verify and me.

This repository keeps its reasoning in `docs/` rather than in comments, so check
there before inferring intent from the code.

- **How to work here**: [CONTRIBUTING.md](CONTRIBUTING.md): the compatibility
  rule, the documentation rules, and what to run before committing.
- **Before writing code**: the [code conventions](docs/architecture/conventions.md).
- **Where things are**: each package's doc comment names its document, and
  [docs/README.md](docs/README.md) lists every one.

Rules for agents:

- When a change cannot stay within the
  [v1 contract](CONTRIBUTING.md#compatibility), stop and ask rather than work
  around it.
- When a document disagrees with the code and it is the code that looks wrong,
  say so rather than change either silently.
- Only commit or push when asked.
