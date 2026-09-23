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
  └─ 10. One transaction: insert the session, insert the refresh token
           └─ Output{UserID, SessionID, AccessToken, RefreshToken}
```

**Every failure is the same response.** Steps 1 to 5 all end in
`401 INVALID_CREDENTIALS`, so nothing tells a malformed username from an
unknown user or a wrong password. Login therefore never answers `422`, even
for input that registration would reject with one. The five errors are
distinct values carrying a different `reason` (`malformed username`,
`malformed password`, `user not found`, `password not found`, `password
mismatch`), which the HTTP layer logs and never sends to the client.

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
