# Login

`internal/usecase/login`

Verifies a username and password, opens a session, and hands back both
tokens.

```
Input{Username, Password}
  │
  ├─ 1. Parse the username      → invalid?   → ErrUsernameMalformed
  ├─ 2. Parse the password      → invalid?   → ErrPasswordMalformed
  ├─ 3. FindByUsername          → not found? → verify against the dummy hash → ErrUserNotFound
  ├─ 4. FindByUserID (password) → not found? → verify against the dummy hash → ErrPasswordNotFound
  ├─ 5. Verify(plain, hash)     → mismatch?  → ErrPasswordMismatch
  │
  ├─ 6. now := clock.Now()
  ├─ 7. Create the Session
  ├─ 8. Issue the access token (JWT: sub, exp)
  ├─ 9. Create the RefreshToken (no parent, expires at now + REFRESH_TOKEN_TTL)
  │       └─ generates the secret; the token keeps its SHA-256
  │
  └─ 10. One transaction:
           ├─ LockHash, only while the stored hash is still the one step 5 checked
           │     └─ changed meanwhile → roll back → ErrPasswordChanged
           ├─ insert the session
           └─ insert the refresh token
                 └─ Output{UserID, SessionID, AccessToken, RefreshToken}
```

**Steps 1 to 5 are the [shared credential check](README.md#the-shared-credential-check)**,
which change password runs too, so the two reject credentials alike.

**Every failure is the same response.** Steps 1 to 5 and 10 all end in
`401 invalid_credentials`, so nothing tells a malformed username from an
unknown user or a wrong password. Login therefore never answers `422`, even
for input that registration would reject with one. The six errors are
distinct values carrying a different `reason` (`malformed username`,
`malformed password`, `user not found`, `password not found`, `password
mismatch`, `password changed`), which the HTTP layer logs and never sends to
the client.

**A missing account costs a bcrypt comparison too.** Bcrypt takes hundreds of
milliseconds, so failing fast in steps 3 and 4 would let response time reveal
whether an account exists. Both steps check the password against a dummy hash
first. `internal/bootstrap` creates that hash at startup with the configured
hasher, so it costs the same as a stored hash; see
[decision 0045](../../development/decisions/0045-login-dummy-hash.md). Steps 1
and 2 do fail fast: their time depends only on the input, and the format rules
are public. Changing `BCRYPT_COST` once there are users reopens the timing
difference for older accounts; see
[Limitations](../../limitations.md#login-timing-after-a-bcrypt_cost-change).

**The first refresh token has no parent**, which marks the start of the
rotation chain. Every later one is created by [refresh](refresh.md).

**Step 10 is atomic**, so a refresh token never exists without its session.

**A password change reaches the session.** [Changing the
password](change-password.md) revokes the sessions of the user that exist
when the change commits. Steps 3 to 5 run before the transaction, and bcrypt
makes them last hundreds of milliseconds: a change that committed in between
would miss the session step 10 is about to save, which would then stay
refreshable for good, opened with a password that no longer works. So step 10
first locks the password, only while its hash is still the one step 5 checked
([guarded writes](../persistence/repositories.md#guarded-writes)). A change
that comes later waits for the transaction to commit, and its revocation
reaches the session. A change that came first leaves the lock nothing to
match, once it commits: the transaction rolls back, the access token from
step 8 is never returned, and the login gets `401 invalid_credentials` with
the reason `password changed`, as it would have a moment later. The lock is
shared, so logins do not wait for each other; see
[decision 0055](../../development/decisions/0055-login-locks-the-password-it-checked.md).
