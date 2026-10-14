# Commits

One change per commit, with a message in
[Conventional Commits](https://www.conventionalcommits.org/) form:

```
type(scope): summary

Body, wrapped at 72 columns.
```

The subject says what changed; the body says why, which is the part the diff
cannot show.

## Subject

`type(scope): summary`, at most 72 characters, no trailing full stop. The
summary is a lowercase imperative phrase that completes "this commit will…",
and names the identifiers it touches rather than the files: *remove unused
`ChangeUsername`*, not *edit user.go*.

### Type

| Type | For |
|---|---|
| `feat` | New behaviour, in the API or in the code behind it |
| `fix` | Behaviour that was wrong and now is not |
| `refactor` | The same behaviour in a different shape: a rename, a move, an extraction |
| `perf` | The same behaviour, faster or lighter |
| `test` | Test code and nothing else |
| `docs` | Documentation and nothing else: `docs/`, the READMEs, comments |
| `style` | Whitespace, import order, `gofmt` output |
| `build` | `go.mod`, `go.sum`, the Dockerfiles, the Compose files |
| `chore` | The repository around the code: `.gitignore`, editor and tooling configuration |
| `revert` | Undoing an earlier commit, named in the body |

The type describes the change, not the files it touched: an endpoint that
arrives with its documentation is `feat`, and `test` and `docs` are for commits
that touch nothing else. A change that both fixes and refactors is two commits;
when it truly cannot be split, the type is the one the reader cares about.

### Scope

Optional: the package the change lives in, left out when it spans several. It
is the package's path under `internal/`, shortened when that is unambiguous:
`session` rather than `domain/session`, but `usecase/login`, since `api` has a
login too. Adapters are `infra/<name>` (`infra/postgres`, `infra/jwt`), test
helpers go by their package name (`porttest`, `postgrestest`), and outside
`internal/` the scopes are `cmd`, `migrations`, `e2e` and `docstest`.

## Body

Worth writing whenever the subject leaves a *why* unanswered. One that is worth
reading says how things stood and what was wrong, in the past tense; what the
change does and why that shape, in the present; and, when there is something to
say, what it costs or deliberately leaves alone. Prose, not bullet lists, and no
*this commit*:

```
refactor(config): reduce env resolution to parsing only

The env type used to know about required variables and defaults, which
duplicated what NewConfig already decided and made LoadConfig look like
it had already validated the values it passed on. Resolve now returns a
nil pointer for a variable that is not set and nothing else: NewConfig is
the single place that says what is required, what the default is and
which values are acceptable.
```

When the reasoning outgrows a body, it belongs in the documentation, and the
message points at it.

## What goes in one commit

One change, complete: the code, its tests, and the documentation it would
otherwise make wrong. Each commit builds and passes its tests on its own, so
that `git bisect` and a revert land somewhere sensible. A rename across thirty
files is one commit; a morning that fixed a bug, extracted a helper and added a
setting is three.
