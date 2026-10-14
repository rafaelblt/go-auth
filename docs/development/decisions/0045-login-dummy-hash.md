# A missing account costs a bcrypt comparison too

- **Status:** Accepted

## Invariant

Every credential check that fails past the format checks pays a bcrypt
comparison, against a hash of the configured cost. The comparison that exists
only for the time it takes is in `credential.Checker.rejectWithDummyVerify`,
and its hash is the one `internal/bootstrap` builds at startup with the
configured hasher. Dropping the call, or hard-coding a hash of a fixed cost,
reopens the timing difference.

## Context

Login and change password answer every credential failure with the same
`401 invalid_credentials`
([Deliberately vague errors](../../api/errors.md#deliberately-vague-errors)),
but the same answer does not take the same time by itself. Returning as soon as
the user or its password row is missing skips the bcrypt comparison, which is
slow by design, so an unknown username would answer hundreds of milliseconds
faster than a wrong password, and response time alone would reveal whether an
account exists.

The miss has to do the work of a real check, and a bcrypt hash carries its
cost: since `BCRYPT_COST` is configurable, a dummy hash of any other cost takes
a different time to verify.

## Decision

When `FindByUsername` finds no user, or `FindByUserID` no password, the
[shared credential check](../../architecture/usecases.md#shared-credential-check)
verifies the submitted password against a dummy hash, ignores the result, and
returns the rejection for that miss. An error from that call is unexpected,
since it can only mean the dummy hash is broken.

The dummy hash is a `password.Hashed` that `internal/bootstrap` produces at
startup by hashing a fixed password with the configured `PasswordHasher`, so it
always has the configured cost. Login and change password receive it in their
`Config` and pass it to their `credential.Checker`, which keeps depending on
`PasswordChecker` alone.

The format checks keep failing fast: their time depends only on the input, and
the format rules are public.

## Alternatives

- **A hash hard-coded in the use case**: it has a fixed cost, so any other
  `BCRYPT_COST` reopens the gap, and another hasher would reject it with an
  error, turning the miss into a `500` that reveals more than timing.
- **Hashing lazily on the first miss**: no startup cost, but the first miss
  pays two hashes, and the use cases gain a dependency used only for this.
- **Padding every response to a fixed duration**: the right duration depends on
  hardware and load, a slow outlier still leaks, and random delays average out.

## Consequences

- Startup hashes one password at the configured cost: roughly 200–400 ms at
  cost 12, far more at the top of the range.
- A miss still skips the `FindByUserID` query, a difference of one database
  round trip, small next to bcrypt and network jitter, and accepted.
- The dummy hash has the configured cost, not the cost of every stored hash, so
  after a `BCRYPT_COST` change the gap reopens for older accounts
  ([Limitations](../../limitations.md#login-timing-after-a-bcrypt_cost-change)).
- This hides account existence from login and change password only:
  registration answers `409` for a taken username
  ([Limitations](../../limitations.md#registration-reveals-whether-a-username-exists)).
