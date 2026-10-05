# Validation

`internal/validation`

The types for input rules, shared by two callers:

- **the domain.** [`Username`](user.md#username) and
  [`Plain`](password.md#plain) use them, and the API reports their failures as
  `422` responses.
- **`internal/config`.** The environment variables are validated with the same
  `Issue`, the same validators and the same `Accumulator`, and the failures
  become the [startup error](../../configuration.md#startup-validation). They
  never reach HTTP.

The package is documented here, next to the domain types that are its main
caller, but it is not a domain package. Anything changed in it — the shape of
`Issue`, of `ValidationError`, of the `Accumulator` — changes the startup error
report too.

## Issue

One broken rule: a `code` and a `details` map.

```go
Issue{ code: "too_short", details: {"min": 3, "unit": "code_point"} }
```

`Issue` does **not** implement `error`. It is data about a failure, and
nothing treats a single issue as an error: code works with `Issues` or
`ValidationError`. It would also be misleading: `details` is a map, so `Issue`
is not comparable, and `errors.Is(err, IssueTooShort(3, UnitCodePoint))` would
quietly return `false`, since `errors.Is` does not compare targets that are
not comparable. A check that always fails without complaining is worse than
no check.

## Validators

A `Validator[T]` is `func(T) *Issue`, where `nil` means the value passed.
Every caller wants the issue, so an `error` would only force an `errors.As`
on each of them, and a `bool` would leave unclear which value means valid.

| Validator            | Issue                | `details`            | Used by |
| -------------------- | -------------------- | -------------------- | ------- |
| `MinLength(n, unit)` | `too_short`          | `min`, `unit`        | domain  |
| `MaxLength(n, unit)` | `too_long`           | `max`, `unit`        | domain  |
| `AllowedChars(set)`  | `invalid_characters` | —                    | domain  |
| `Required[T]()`      | `required`           | —                    | config  |
| `Positive[T]()`      | `not_positive`       | —                    | config  |

The first three are the ones whose failures a client sees; they are the whole
[field code catalog](../../api/errors.md#field-codes) of the API. The
last two exist for `internal/config`: `Required` catches a missing
`DATABASE_URL` or `ADDRESS`, and `Positive` catches a cost or a TTL that is zero
or negative.

`internal/config` also declares a validator of its own rather than using a
generic one: `allowedLogFormat` checks `LOG_FORMAT` against the accepted values
and returns `IssueNotAllowed`, whose code is `not_allowed` and whose `details`
carry `allowed`. A validator is just a `func(T) *Issue`, so a package can write
one without this one knowing about it — which is why the table above lists what
this package exports, not every validator in the codebase.

`Validate(value, validators...)` runs every validator and collects every
failure, so one call reports everything wrong with a value.

`MinLength` and `MaxLength` are separate, so each validator produces one code.
A single `Length(min, max)` would take two adjacent `int`s, which invites
swapping them, and nothing would reject `Length(32, 3)`.

Every length is counted in an explicit `LengthUnit`, `code_point` or `byte`,
and the unit reaches the API response. Go's `len` counts bytes, JavaScript's
`.length` counts UTF-16 code units, and a user counts characters: for
`"josé"`, that is 5, 4 and 4. With the unit left implicit, client and server
would eventually disagree.

## Accumulator

Collects the failures of several fields:

```go
acc := validation.NewAccumulator()

username, issues := user.NewUsername(input.Username)
acc.Add(FieldUsername, issues)

plain, issues := password.NewPlain(input.Password)
acc.Add(FieldPassword, issues)

if err := acc.Err(); err != nil {
    return Output{}, err
}
```

`Add` returns nothing, which is what keeps this readable: an error return
would need an `if err != nil` after every field. `Err()` returns a
`ValidationError` holding every `FieldError`, or `nil` if every field passed.
This is how registration reports a bad username and a bad password in one
response.

`internal/config` uses it the same way, over seven fields instead of two, which is
how a startup failure lists every misconfigured variable at once instead of the
first one.
