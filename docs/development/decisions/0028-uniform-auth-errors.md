# Uniform errors on login and refresh

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication, api
- **Related:** [0030](0030-kind-based-status-mapping.md)

## Context

Login and refresh can each fail for several reasons. Distinguishing them would
tell an attacker whether an account exists, or whether a stolen token has
already tripped reuse detection.

## Decision

Login returns `INVALID_CREDENTIALS` for every failure; refresh returns
`INVALID_TOKEN` for all four of its failure modes.

`UseCaseError.reason` carries the specific cause into the logs only. The
`kind`/`code`/`reason` split exists precisely so an error can be precise
internally and vague externally. How `kind` becomes a status code is covered in
[kind-based status mapping](0030-kind-based-status-mapping.md).
