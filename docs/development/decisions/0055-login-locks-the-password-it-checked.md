# Login locks the password it checked until its session is saved

- **Status:** Accepted
- **Date:** 2026-10-10
- **Related:** [Guarded writes](../../architecture/persistence/repositories.md#guarded-writes), [Login](../../architecture/usecases/login.md), [Change password](../../architecture/usecases/change-password.md), [0050](0050-refresh-token-use-is-settled-at-write.md)

## Invariant

Login saves a session only in a transaction that first ran
`PasswordWriter.LockHash` on the password it checked, which locks the
`passwords` row `FOR SHARE` only while its hash is still the one checked, and a
password change updates that row in the same transaction as it revokes the
sessions.

## Context

[Changing the password](../../architecture/usecases/change-password.md)
revokes every session of the user, because sessions have no maximum age and a
refresh token obtained with the old password would otherwise keep its holder
signed in for good. The revocation is an `UPDATE` of the sessions that exist
when it runs.

[Login](../../architecture/usecases/login.md) reads the password and compares
it with bcrypt before its unit of work opens, `port.UowDeps` holding only
writers, and only then inserts the session and its refresh token. A change
that commits between the two revokes the sessions that exist, and the one
login inserts a moment later is not among them. It works, it refreshes, and
each refresh resets its expiry, although the password it was opened with no
longer does. Of 20 changes made while one client kept logging in with the old
password, all 20 left such a session behind.

The rule, that no session opened with a password outlives a change of it,
spans two tables: login writes `sessions` and the change writes `passwords`.
[Guarded writes](../../architecture/persistence/repositories.md#guarded-writes)
settle a rule on the row being written, and here neither write can see the
other. Under `READ COMMITTED` the revocation does not see a session that is
not committed yet, and a condition on the session insert does not see a hash
that is not committed yet.

## Decision

Login's unit of work first runs `PasswordWriter.LockHash` with the password it
checked, then inserts the session and the refresh token. `PasswordRepo` takes
the lock on the row both requests depend on:

```sql
SELECT 1 FROM passwords WHERE id = @id AND hash = @hash FOR SHARE
```

Whichever comes first, the two meet there:

- **The change first.** Its `UPDATE` holds the row. The lock waits for the
  commit, checks the condition again against the new row and matches nothing,
  so `LockHash` returns `password.ErrHashChanged`. Login rolls back and answers
  `ErrPasswordChanged`: `401 invalid_credentials`, with the reason
  `password changed` in the log. The access token it issued is never returned.
  It is the answer the login would have had a moment later.
- **The login first.** The change's `UPDATE` waits for login to commit. The
  revocation that follows runs with a fresh snapshot, sees the session and
  revokes it.

**`FOR SHARE`, the weakest lock that works.** The change's
`UPDATE passwords SET hash` takes `FOR NO KEY UPDATE`, which conflicts with
`FOR SHARE` and not with `FOR KEY SHARE`, so a weaker lock would not make
either wait. A stronger one, `FOR UPDATE`, would make the logins of one user
wait for each other; shared locks do not.

**A guarded statement, on the writer.** `LockHash` has the shape of a guarded
write: the condition in the `WHERE`, the outcome read from `RowsAffected`, the
sentinel from the domain, and the existence check only when nothing matched.
It is on `PasswordWriter` because it means something only inside a
transaction, and it returns no data, only whether the row was still as login
read it, so `UowDeps` still holds no reader.

**Bcrypt stays outside the transaction.** The lock is taken after the
comparison, so the transaction holds it for one indexed query and two inserts,
not for the hundreds of milliseconds of a bcrypt comparison.

**Change password is unchanged.** Its guarded `UpdateHash` already updates
the row in the same transaction as the revocation, which is what the lock
waits for.

## Alternatives considered

- **A guarded session insert**, inserting only while the stored hash is still
  the one checked, which Limitations proposed. It does not close the window:
  with both transactions in flight, the condition does not see the
  uncommitted change, and the revocation does not see the uncommitted session.
  Tried against two connections, it left a session active.
- **Reading inside the unit of work with `SELECT … FOR UPDATE`, as a rule for
  every write use case.** Correct here, but either bcrypt runs inside the
  transaction, holding a pool connection and an exclusive lock for hundreds of
  milliseconds per login, so that wrong passwords sent for one username make
  its logins wait in line, or the password is read twice and the second read
  is this lock. It needs readers in `UowDeps`, hides the lock behind a read,
  and still leaves guarded writes for rules about rows that do not exist yet,
  such as a free username
  ([0048](0048-duplicate-username-reported-by-the-writer.md)). Two patterns
  for one job. [0050](0050-refresh-token-use-is-settled-at-write.md) weighed it
  too.
- **A security stamp**: a version on the user, copied into each session at
  login and compared on every refresh. It closes the window without a lock,
  but needs a migration, a column on `sessions` and a check in refresh. It is
  the better tool if the service needs a general way to sign a user out
  everywhere, and worth revisiting then.
- **`SERIALIZABLE` transactions.** PostgreSQL detects the conflict only if
  login reads the password inside its transaction, which brings bcrypt back
  in, and every serialization failure needs a retry.
- **A lock in memory, per user.** The service can run as several instances
  over one database, and they share nothing else.
- **Keeping it as a documented limitation.** The window is about one bcrypt
  comparison long and needs the old password. But what it yields is a session
  that never ends, held by whoever the change was most likely meant to lock
  out.

## Consequences

- No session opened with a password outlives a change of it. Access tokens
  already issued stay valid until they expire, as after any revocation.
- A login that loses the race is answered `401` for a password that was
  correct when it was sent.
- Every login runs one more indexed query. A password change waits for the
  logins of the same user that are saving their sessions, for milliseconds.
- Any future path that opens a session from a checked password has to take the
  same lock, or it reopens the window.
