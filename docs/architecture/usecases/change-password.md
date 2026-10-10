# Change password

`internal/usecase/changepassword`

Replaces a user's password, given the username and the current one, and
revokes every session of the user. It returns no token: the user logs in again
with the new password.

```
Input{Username, CurrentPassword, NewPassword}
  │
  ├─ 1. Parse the username          → invalid?   → ErrUsernameMalformed
  ├─ 2. Parse the current password  → invalid?   → ErrPasswordMalformed
  ├─ 3. Validate the new password   → invalid or the current one? → ValidationError → 422
  ├─ 4. FindByUsername              → not found? → verify against the dummy hash → ErrUserNotFound
  ├─ 5. FindByUserID (password)     → not found? → verify against the dummy hash → ErrPasswordNotFound
  ├─ 6. Verify(current, hash)       → mismatch?  → ErrPasswordMismatch
  │
  ├─ 7. Hash the new password (bcrypt)
  ├─ 8. now := clock.Now(), Password.ChangeHash(new hash, now)
  │
  └─ 9. One transaction:
          ├─ UpdateHash, only while the stored hash is still the one step 6 checked
          │     └─ replaced meanwhile → roll back → ErrPasswordChanged
          └─ RevokeAllByUserID(now)
                └─ Output{User}
```

**The credentials are checked as login checks them.** Steps 1, 2 and 4 to 6
are the [shared credential check](README.md#the-shared-credential-check) that
[login](login.md) runs too, and end in the same `401 invalid_credentials`,
with the same `reason`s in the log. A missing account costs a bcrypt comparison
against the dummy hash here too: otherwise this endpoint would tell an unknown
username from a wrong password by its response time, which login takes care
not to do
([decision 0045](../../development/decisions/0045-login-dummy-hash.md)).
`internal/bootstrap` creates one dummy hash at startup and gives it to both.

**The new password follows the [registration rules](../domain/password.md#plain),
and is not the current one**, and a failure is a `422` naming `new_password`.
A change to the same password would change nothing but still sign the user out
everywhere, so it is refused with `same_as_current`. Only the current password
is compared: the service keeps no history, and a password used before it can
come back. It is all checked before the account is looked up, against
`current_password` as sent: the answer depends only on the input and on rules
that are public, so it says nothing about the account, and a request that
cannot succeed costs no bcrypt comparison.

**Every session of the user is revoked.** A password is most often changed
because someone else may know it, and sessions have
[no maximum age](../../limitations.md#sessions-have-no-maximum-age): a refresh
token obtained with the old password would otherwise keep its holder signed in
for good. The cost is that the user is signed out everywhere, including where
the password was changed. Access tokens already issued stay valid until they
expire, as after any revocation
([Limitations](../../limitations.md#access-tokens-cannot-be-revoked)).

**Step 9 is atomic**, so the password never changes without the sessions being
revoked.

**The current password is all it takes.** Whoever knows it can change it, and
lock the owner out, with no reset to recover from; see
[Limitations](../../limitations.md#whoever-knows-the-password-can-lock-the-owner-out).

**Concurrent changes.** Two requests with the same current password can both
pass step 6. The write settles it: `UpdateHash` replaces the hash only while it
is still the one step 6 checked, so exactly one applies
([guarded writes](../persistence/repositories.md#guarded-writes)). The other
rolls back and gets `401 invalid_credentials`, with the reason
`password changed`: by the time it wrote, the password it presented was no
longer the user's, and it is answered as it would have been a moment later.
Without the guard both would succeed, the later write would win, and one caller
would hold a password that does not work.

**Logins running alongside.** The revocation reaches the sessions that exist
when the transaction commits, and a login that checked the old password just
before may not have saved its session yet. [Login](login.md) saves it only
while it holds a lock on the password, taken while the hash is still the one it
checked. Either step 9 waits for that login to commit, and the revocation
reaches its session, or the login finds the new hash and is rejected. No
session opened with the old password outlives the change.
