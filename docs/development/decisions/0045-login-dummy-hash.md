# Login verifies a dummy hash when the account is missing

- **Status:** Accepted
- **Date:** 2026-09-13
- **Related:** [Deliberately vague errors](../../api/errors.md#deliberately-vague-errors)

## Invariant

Every login failure past the format checks pays a bcrypt comparison, against a
hash of the configured cost. The calls that exist only for the time they take are
in `rejectWithDummyVerify`, and the hash comes from
`login.Config.DummyPasswordHash`, which `internal/bootstrap` builds with the
configured hasher. Dropping either call, or hard-coding a hash of a fixed cost,
reopens the timing difference.

## Context

[Uniform errors](../../api/errors.md#deliberately-vague-errors) make every login failure return
the same response, but not take the same time. Login used to return as soon as
the user or its password row was missing, before the bcrypt comparison.
Bcrypt is slow by design, so an unknown username answered hundreds of
milliseconds faster than a wrong password, and response time alone revealed
whether an account existed.

The miss paths have to do the same work as a real check. The obstacle is that
a bcrypt hash carries its cost, and `BCRYPT_COST` is configurable: a dummy
hash of any other cost takes a different time to verify.

## Decision

When `FindByUsername` finds no user, or `FindByUserID` finds no password,
login calls `PasswordChecker.Verify` with the submitted password and a dummy
hash, ignores the result, and returns the rejection for that miss
(`ErrUserNotFound` or `ErrPasswordNotFound`, both answered as
`INVALID_CREDENTIALS`). An error from that call is returned as an unexpected
error, since it can only mean the dummy hash is broken.

The dummy hash is a `password.Hashed` injected through
`login.Config.DummyPasswordHash`. `internal/bootstrap` produces it at startup
by hashing a fixed password with the configured `PasswordHasher`, so it always
has the configured cost, and `Login` keeps depending on `PasswordChecker`
alone.

The validation early returns (username and password format) are left as they
are: their timing depends only on the input, and the format rules are public.

## Alternatives considered

- **A hard-coded hash and password in the `login` package** — the hash would
  have a fixed cost, so any other `BCRYPT_COST` reopens the gap in one
  direction or the other. It would also put a bcrypt string inside a use case,
  and a different hasher would reject it with an error, turning the miss path
  into a `500` that reveals the account is missing more plainly than timing.
- **Injecting `PasswordHasher` and hashing lazily on the first miss** — no
  startup cost, but the first miss pays for two hashes, and login gains a
  dependency it uses only for this.
- **A dummy hash owned by the adapter**, exposed through a new port method —
  the adapter knows its cost, but the port would carry a method that exists
  only for one use case, and whether to return early is the use case's policy.
- **Padding responses to a fixed minimum duration** — the right duration
  depends on hardware and load, and a slow outlier still leaks. Random delays
  only add noise that averages out over repeated requests.

## Consequences

- Startup hashes one password at the configured cost: roughly 200–400 ms at
  cost 12, and far more at the upper end of the range.
- A miss still skips the `FindByUserID` query. The difference is one database
  round trip, small next to bcrypt and to network jitter, and accepted.
- The dummy hash matches the configured cost, not the cost of every stored
  hash. After a `BCRYPT_COST` change the gap reopens for accounts hashed
  before it ([Limitations](../../limitations.md#login-timing-after-a-bcrypt_cost-change)).
- This removes the timing signal from login only. Register still answers
  `409 USERNAME_ALREADY_EXISTS` for a taken username, so account existence is
  not secret; what limits enumeration through either endpoint is rate limiting.
