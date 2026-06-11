# AGENTS.md

## Read First

This repository contains `nobl9-go`, the Go SDK for Nobl9.
Do not use this file as a replacement for project documentation.

Read the existing docs before changing behavior, tests, generated code, or
release logic:

- [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for development workflow,
  Makefile behavior, CI, validation tests, e2e tests, code generation, and
  dependencies.
- [README.md](./README.md) for user-facing SDK purpose, package layout, and
  usage.
- [docs/CONTRIBUTING.md](./docs/CONTRIBUTING.md) for contribution workflow.
- [docs/RELEASE.md](./docs/RELEASE.md) for release process details.

If a workflow is documented there, follow the existing doc instead of adding
a second version here.

## Command Policy

Use the Makefile targets instead of calling tools directly.
To inspect available targets, run: `make help`.
The CI workflows under [.github/workflows](./.github/workflows/) use the same
Makefile targets or equivalent commands, so treat them as the local source of
verification commands.

## Testing Policy

Prefer tests at the level where behavior is exposed.
For manifest validation, test the whole object and use `internal/testutils` as
described in
[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md#govy-validation-tests).

For SDK client or endpoint behavior visible through the platform API, add or
update e2e coverage under [tests](./tests/) unless there is a concrete reason
not to.
If e2e coverage is not practical, state the reason in the PR or handoff notes.

End-to-end tests talk to the Nobl9 platform API.
Do not run them without explicit user permission.

Before writing or modifying e2e tests, read:

- [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md#end-to-end-test)
- [tests/e2etestutils](./tests/e2etestutils/)
- sample existing tests to follow established object setup, cleanup, and retry
  patterns

## Change Policy

Follow existing package layout, endpoint patterns, manifest object patterns,
and test style before adding new abstractions.

Do not edit generated files directly.
If generated output is stale, update source definitions and run
`make generate`, then verify with `make check/generate`.

## Verification

Always verify changes with project targets before claiming completion.
For Markdown-only changes, run `make check/markdown` at minimum.

If a command cannot be run locally, report the exact command and exact error.
Do not replace failed verification with assumptions.
