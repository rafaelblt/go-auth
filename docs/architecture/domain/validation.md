# Validation

`internal/validation`

The types for input rules. [`Username`](user.md#username) and
[`Plain`](password.md#plain) use them, and the API reports their failures as
`422` responses.

## Issue

One broken rule: a `code` and a `details` map.

```go
Issue{ code: "TOO_SHORT", details: {"min": 3, "unit": "code_point"} }
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

| Validator            | Issue                |
| -------------------- | -------------------- |
| `MinLength(n, unit)` | `TOO_SHORT`          |
| `MaxLength(n, unit)` | `TOO_LONG`           |
| `AllowedChars(set)`  | `INVALID_CHARACTERS` |

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
