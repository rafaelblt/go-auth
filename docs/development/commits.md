# Commits

One change per commit, with a message in
[Conventional Commits](https://www.conventionalcommits.org/) form:

```
type(scope): summary

Body, wrapped at 72 columns.
```

The subject says what changed. The body says why, and that is where the
value is: the diff already shows the change, and the message is the only
place the reasoning survives.

## Subject

`type(scope): summary`, at most 72 characters, no trailing full stop.

The summary is a lowercase imperative phrase that completes "this commit
will…": *close the pool when startup fails*, *unexport DTO fields*. It
names the identifiers it touches, not the files: *remove unused
`ChangeUsername`* rather than *edit user.go*.

### Type

Only these, because they are the ones a reader already knows.

| Type | For |
|---|---|
| `feat` | New behaviour, in the API or in the code behind it. |
| `fix` | Behaviour that was wrong and now is not. |
| `refactor` | The same behaviour in a different shape: a rename, a move, an extraction, a signature. |
| `perf` | The same behaviour, faster or lighter. |
| `test` | Test code and nothing else: filling a gap in coverage, reworking the helpers in `internal/testutil`. |
| `docs` | Documentation and nothing else: `docs/`, the READMEs, doc comments, decision records. |
| `style` | Whitespace, import order, `gofmt` output. Nothing a reader has to think about. |
| `build` | How the project is built and what from: `go.mod`, `go.sum`, the Dockerfiles, the Compose files. |
| `chore` | The repository around the code: `.gitignore`, editor and tooling configuration. |
| `revert` | Undoing an earlier commit, named in the body. |

The type describes the change, not the files it touched. A rename that
drags its tests along is `refactor`, and an endpoint that arrives with its
documentation is `feat`; `test` and `docs` are for commits that touch
nothing else.

A change that both fixes and refactors is two commits. When it genuinely
cannot be split, the type is the one the reader cares about: `fix` over
`refactor`, `feat` over `fix`, and anything over `chore`.

### Scope

The scope is the package the change lives in, written as its path under
`internal/`, and it is optional. Use it when the change sits in one place,
leave it out when it spans several packages or the repository as a whole.

| Scope | Package |
|---|---|
| `user`, `password`, `session` | `internal/domain/*` |
| `validation`, `shared`, `port` | `internal/validation`, `internal/shared`, `internal/port` |
| `usecase`, `usecase/register`, `usecase/login`, `usecase/refresh` | `internal/usecase/...` |
| `api` | `internal/api` |
| `infra/postgres`, `infra/bcrypt`, `infra/jwt`, `infra/migrate` | `internal/infra/*` |
| `bootstrap`, `config`, `cmd` | `internal/bootstrap`, `internal/config`, `cmd` |
| `testutil`, `apitest`, `porttest`, `postgrestest`, `usertest`, `passwordtest`, `sessiontest`, `migratetest` | `internal/testutil/...` |
| `migrations`, `e2e` | `migrations`, `tests/e2e` |

Write the shorter form when it is unambiguous: `session`, not
`domain/session`. Write the longer form when the short one is not:
`usecase/login` distinguishes the use case from `api`, which also has a
login path.

## Body

Optional, wrapped at 72 columns, and worth writing whenever the subject
leaves a *why* unanswered. A change whose reason is plain from its
subject — removing dead code, adding a test for an existing function,
accepting a decision record — does not need one.

A body that is worth reading follows the shape of the decision it records:

1. **How it stood, and what was wrong with it.** Past tense, naming the
   functions and types involved. This is the part a reader cannot recover
   from the diff.
2. **What the change does, and why that shape.** Present tense, describing
   the new state as it now is.
3. **What it costs, or what was deliberately left alone.** Only when there
   is something to say.

Prose, not bullet lists, and no *this commit* or *I changed*. From the
history:

```
refactor(config): reduce env resolution to parsing only

The env type used to know about required variables and defaults, which
duplicated what NewConfig already decided and made LoadConfig look like
it had already validated the values it passed on. Resolve now returns a
nil pointer for a variable that is not set and nothing else: NewConfig is
the single place that says what is required, what the default is and
which values are acceptable.
```

When the reasoning outgrows a body — several packages, real alternatives
to weigh — it belongs in a [decision record](decisions/README.md)
instead, and the commit message points at it.

## What goes in one commit

A commit is one change, complete: the code, the tests written with it, and
the documentation it makes wrong if left behind. It should build and pass
its tests on its own, so that `git bisect` and a revert both land somewhere
sensible.

That cuts both ways. A rename that touches thirty files is one commit, and a
morning's work that happens to fix a bug, extract a helper and add a
configuration option is three.
