# Minimum password length

- **Status:** Accepted
- **Date:** 2026-09-17
- **Areas:** domain
- **Related:** [0011](0011-password-input-rules.md)

## Context

The minimum was 4 code points, set by
[Password input rules](0011-password-input-rules.md). That was a floor on "is
this a password at all", not a security recommendation.

Worth settling before 1.0, since raising it later is a policy change existing
users would feel at their next password change. Raising it now costs nothing.

## Decision

A password is at least 8 code points, the length NIST SP 800-63B requires of
a memorised secret the user chooses.

**Eight is the smallest defensible default.** The service is cloned and run by
whoever deploys it, so the default is what nearly every deployment gets. A
default that accepts `1234` is a liability in a project whose purpose is to be
reused.

**Four code points describes a PIN, and this service is not built for one.** A
short numeric secret is defensible only with the controls that make its tiny
search space unreachable: attempts counted per account, lockout, rate limiting
per caller, usually a second factor. This service has none of them, so a
4 code point minimum offered the unsafe half of that design — a handful of
possibilities with unlimited guesses, bounded only by the cost of a bcrypt
comparison. A PIN mode is a decision to take together with those controls, not
a minimum to lower in advance of them.

**A lower minimum is one constant away.** `password.PlainMinCodePoints` is a
compile-time constant in a repository every operator already clones. Keeping
the default at 4 would not have bought flexibility that 8 removes; it would
only have moved the edit from the rare deployment to the common one.

**The minimum is not configurable.** The validators are a package-level slice
built from the constants, so a runtime policy means either threading it
through `NewPlain` and every caller, or package-level mutable state. That is
real coupling for a knob almost nobody turns and that editing the constant
already answers. Deployment-shaped values — token lifetimes, database
settings — belong in `internal/config`; the strength of a password the service
accepts is a property of the service.

This is a floor, not a strength policy. Nothing here restricts characters or
requires composition — [0011](0011-password-input-rules.md) decided that, and
a length minimum is precisely the rule that survives when composition rules
are dropped.

## Alternatives considered

- **Keep 4 code points** — see above: it serves a PIN design the service does
  not implement, in the position that most deployments inherit by default.
- **Make the minimum configurable** — a policy value injected into the domain,
  at the cost described above, so that a deployment can lower the floor
  without the protections that would justify lowering it.

## Consequences

- Passwords shorter than 8 code points are rejected at registration, with the
  `TOO_SHORT` issue the rule already produced at 4.
- Existing stored passwords are unaffected: the rule applies to input, and
  nothing revalidates a hash. A user whose password predates this record keeps
  signing in until they change it.
- Revisit if the service gains per-account attempt limiting and lockout, which
  is what a deliberate short-secret mode would need.
