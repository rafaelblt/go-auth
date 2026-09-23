# Register

`internal/usecase/register`

Creates an account. It does not log the user in: it creates no session and no
token.

```
Input{Username, Password}
  │
  ├─ 1. Validate both fields, collecting every issue
  │       └─ any failure → ValidationError → 422
  │
  ├─ 2. ExistsByUsername?
  │       └─ yes → ErrUsernameAlreadyExists → 409
  │
  ├─ 3. Hash the password (bcrypt)
  ├─ 4. Read the clock once → now
  ├─ 5. Build the User (active, created_at = updated_at = now)
  ├─ 6. Build the Password (referring to the new user's ID)
  │
  └─ 7. One transaction: insert the user, insert the password
          ├─ username taken meanwhile → ErrUsernameAlreadyExists → 409
          └─ Output{User}
```

**Both fields are validated before either is reported**, so a request with a
bad username *and* a bad password gets both errors at once. The rules
themselves are in [Username](../domain/user.md#username) and
[Plain](../domain/password.md#plain).

**Steps 5 and 6 share one `now`**, so the user and its password have the same
timestamps rather than ones microseconds apart.

**Step 7 is atomic.** Without it, a failure between the two inserts could
leave a user without a password: an account nobody can log into, holding its
username forever.

**Step 2 does not decide whether the username is free.** Two concurrent
registrations of the same username can both pass it. The insert in step 7
decides, and the loser gets the same `409` as a registration of a username
taken long ago. Step 2 stays because it answers the common case with an index
lookup, before step 3 spends a bcrypt hash on a username that is taken. See
[decision 0048](../../development/decisions/0048-duplicate-username-reported-by-the-writer.md).
