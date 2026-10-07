# Verify

`internal/usecase/verify`

Checks an access token for a service that would rather not verify it itself,
and says whose it is. It makes the checks a local verifier makes, and nothing
else.

```
Input{AccessToken}
  │
  ├─ 1. Read the clock once → now
  │
  ├─ 2. Validate: EdDSA, a published kid, the signature, exp, sub
  │       ├─ malformed, another alg, unknown kid, bad signature → ErrTokenInvalid → 401
  │       ├─ past its exp                                      → ErrTokenExpired → 401
  │       └─ anything else                                     → 500
  │
  └─ Output{UserID, ExpiresAt, ExpiresIn = ExpiresAt − now}
```

The two `401` paths share one response, `invalid_token`, and differ in their
`reason` (`invalid token`, `token expired`), which only the log sees.

**It reads no database.** Step 2 is `infra/jwt.AccessTokenService.Validate`,
the code a local verifier is pointed to as a reference
([What to verify](../../api/token-verification.md#what-to-verify)), so a token
passes here exactly when it would pass a correct verifier in another service.
That is what lets a service switch between the two without its users noticing.
It also means an access token is no more revocable here than there: it names a
user, not a session, so there is no session to look up, and the tokens of a
revoked session pass until their `exp`
([Revocation is not immediate](../../api/token-verification.md#revocation-is-not-immediate)).

**Step 1 comes before step 2**, so a token that passes had not expired at
`now`, and `ExpiresIn` is not negative.

It is never rate limited ([why](../http.md#rate-limiting)).
