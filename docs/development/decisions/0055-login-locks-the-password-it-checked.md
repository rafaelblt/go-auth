# Login locks the password it checked until its session is saved

- **Status:** Accepted

## Invariant

Login saves a session only in a transaction that first ran
`PasswordWriter.LockHash` on the password it checked, which locks the
`passwords` row `FOR SHARE` only while its hash is still the one checked, and a
password change updates that row in the same transaction as it revokes the
sessions.

## Context

[Changing the password](../../architecture/usecases.md#change-password) revokes
every session of the user, with an `UPDATE` of the sessions that exist when it
runs. [Login](../../architecture/usecases.md#login) compares the password with
bcrypt before its unit of work opens, and only then inserts the session. A
change that commits between the two misses the session login inserts a moment
later, which then works and refreshes for good, although the password it was
opened with no longer does. Of 20 changes made while one client kept logging in
with the old password, all 20 left such a session behind.

The rule, that no session opened with a password outlives a change of it, spans
two tables: login writes `sessions` and the change writes `passwords`. A
[guarded write](../../architecture/persistence.md#guarded-writes) settles a rule
on the row it writes, and here neither write can see the other: under
`READ COMMITTED` the revocation does not see a session not yet committed, and a
condition on the session insert does not see a hash not yet committed.

## Decision

Login's unit of work first locks the password row both requests depend on, then
inserts the session and the refresh token:

```sql
SELECT 1 FROM passwords WHERE id = @id AND hash = @hash FOR SHARE
```

Whichever comes first, the two meet there:

- **The change first.** Its `UPDATE` holds the row, so the lock waits for the
  commit, checks the condition again against the new row and matches nothing.
  Login rolls back, never returns the access token it issued, and answers `401`
  with the reason `password changed`, the answer it would have had a moment
  later.
- **The login first.** The change's `UPDATE` waits for login to commit, and the
  revocation that follows sees the session.

**`FOR SHARE` is the weakest lock that works.** The change's `UPDATE` takes
`FOR NO KEY UPDATE`, which conflicts with `FOR SHARE` but not with
`FOR KEY SHARE`; `FOR UPDATE` would also make the logins of one user wait for
each other, which shared locks do not.

**It is a guarded statement on the writer:** the condition in the `WHERE`, the
outcome from `RowsAffected`, `password.ErrHashChanged` when nothing matched. It
belongs to `PasswordWriter` because it means something only inside a
transaction, and it returns no data, so `UowDeps` still holds no reader. Bcrypt
stays outside the transaction, which holds the lock for one indexed query and
two inserts.

## Alternatives

- **A guarded session insert**, inserting only while the stored hash is still
  the one checked. It does not close the window: with both transactions in
  flight, the condition does not see the uncommitted change, and the revocation
  does not see the uncommitted session. Tried against two connections, it left
  a session active.
- **Reading the password inside the unit of work with `FOR UPDATE`.** Either
  bcrypt runs inside the transaction, holding a connection and an exclusive
  lock for hundreds of milliseconds per login, so that wrong passwords sent for
  one username make its logins queue, or the password is read twice, and the
  second read is this lock.
- **A security stamp**: a version on the user, copied into each session at login
  and compared on every refresh. It closes the window without a lock, but needs
  a migration, a column on `sessions` and a check in refresh. It is the better
  tool if the service needs a general way to sign a user out everywhere.
- **`SERIALIZABLE`**: PostgreSQL detects the conflict only if login reads the
  password inside its transaction, which brings bcrypt back in, and every
  serialization failure needs a retry.
- **Keeping it as a documented limitation**: the window is about one bcrypt
  comparison long and needs the old password, but what it yields is a session
  that never ends, held by whoever the change was most likely meant to lock out.

## Consequences

- No session opened with a password outlives a change of it. Access tokens
  already issued stay valid until they expire, as after any revocation.
- A login that loses the race is answered `401` for a password that was correct
  when it was sent.
- Every login runs one more indexed query, and a change waits, for
  milliseconds, for the logins of the same user that are saving their sessions.
- Any future path that opens a session from a checked password has to take the
  same lock, or it reopens the window.
