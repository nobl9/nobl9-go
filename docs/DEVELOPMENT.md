# Development

This document describes the intricacies of nobl9-go development workflow.
If you see anything missing, feel free to contribute :)

## Pull requests

[Pull request template](../.github/pull_request_template.md)
is provided when you create new PR.
Section worth noting and getting familiar with is located under
`## Release Notes` header.

## Makefile

Run `make help` to display short description for each target.
The provided Makefile will automatically install dev dependencies if they're
missing and place them under `bin`
(this does not apply to `yarn` managed dependencies).
However, it does not detect if the binary you have is up to date with the
versions declaration located in Makefile.
If you see any discrepancies between CI and your local runs, remove the
binaries from `bin` and let Makefile reinstall them with the latest version.

## CI

Continuous integration pipelines utilize the same Makefile commands which
you run locally. This ensures consistent behavior of the executed checks
and makes local debugging easier.

The e2e workflow retries the test target because platform-backed tests can be
transient.

## Modifying objects

> [!CAUTION]
> Validation rule tests are mandatory.
> If you add new fields with validation rules or modify existing ones
> you must cover them with [validation tests](#govy-validation-tests).

When modifying manifest objects, make sure the SDK, generated artifacts, tests,
and examples stay in sync.
At minimum, check the following:

- Update the object type, validation rules, and any shared helpers used by that
  object.
- Update or add object examples under `manifest/v1alpha/examples`.
  These examples feed generated YAML test fixtures and object documentation.
- Run `make generate` after changing object shapes, generated examples,
  validation-derived docs, or generated object implementations.
- Add or update Govy validation tests when validation behavior changes.
  Follow [Govy validation tests](#govy-validation-tests): test whole manifest
  objects, assert the exact error count, and use full govy property paths.
- Add or update end-to-end coverage under [tests](../tests/) when the change
  affects behavior visible through the Objects API.
- Keep [docs/mock_example](./mock_example) in sync when SDK interfaces used by
  the mock example change.

## Testing

### Unit tests

**AVOID** adding unit tests that cover what end-to-end tests are better
equipped to cover, don't mock the HTTP stack, unless it's absolutely necessary.

#### Govy validation tests

Govy validation tests are used as regression material for dependent tools, so
write them against the same object shape users submit.
Do not test an isolated nested struct when the rule is reached through a
manifest object.

When writing validation for Nobl9 manifest objects, follow these rules:

- Use `validation` package ([see](#validation)).
- **ALWAYS** test the whole object and not only its specific fields.
  *TIP*: Create a valid object once and then just modify its specific fields
  to validate them.
- **ALWAYS** use `testutils` package and its `AssertNoError` and
  `AssertContainsErrors`. It not only makes it easier to validate the whole
  object but also it allows recording these tests.
  Recorded tests are planned to be used for regression and dependent
  tools (sloctl, Terraform provider) testing.
- Pass the whole object to `AssertNoError` and `AssertContainsErrors`, not only
  the nested value being changed.
- Assert the exact number of validation errors.
  This catches accidental extra errors, missing cascade stops, and duplicate
  rule execution.
- Use full govy property paths from the object root in `ExpectedError.Prop`,
  for example `metadata.name`,
  `spec.objectives[0].rawMetric.query.prometheus.promql`, or
  `spec.objectives[0].countMetrics.total.appDynamics.metricPath`.
- Prefer matching by `ExpectedError.Code`.
  Use `Message` or `ContainsMessage` only when the rule has no stable error
  code or the message itself is the behavior under test.
- When reusing expected errors from a nested validator, keep the local expected
  paths relative to that validator and use `testutils.PrependPropertyPath` at
  the call site to produce the full object paths.
- If one invalid value is validated in multiple branches, assert each produced
  path and the full error count.

### Recording tests

If you wish to record the tests run `make test/record`.
By default, the tests are recorded inside `./bin` folder.

### End-to-end test

Tests which are run against Nobl9 API are located under [tests](../tests)
folder.
They are standard Go tests, annotated with build tag `e2e_test`, they can
be executed with `make test/e2e`.
In order to run them, you either need to have `~/.config/nobl9/config.toml`
or a set of basic Nobl9 credentials via environment variables:

- *NOBL9_SDK_CLIENT_ID*
- *NOBL9_SDK_CLIENT_SECRET*

There's also a [dispatch action](https://github.com/nobl9/nobl9-go/actions/workflows/e2e-tests-dispatch.yml)
available.

Use `tests/e2etestutils` for generated names, common labels, apply/delete,
and cleanup.
Use the retry helpers from `tests/helpers_test.go` for eventually consistent
API reads instead of adding fixed sleeps in individual tests.

#### Endpoints

All [endpoints](../sdk/endpoints) must follow existing patterns and must implement
end-to-end tests which will cover their interaction with the API.
These tests are not meant to exhaustively test the underlying platform, but rather
ensure the integration works.
It is fine to cover specific platform behaviors when they matter to the SDK
contract, but keep in mind their primary purpose (integration).

## Releases

Refer to [RELEASE.md](./RELEASE.md) for more information on release process.

## Code generation

Some parts of the codebase are automatically generated.
We use the following tools to do that:

- [go-enum](https://github.com/abice/go-enum)
  which is a simple enum generator. We recommend using it instead of writing
  your own const-based enums. It can generate methods for decoding the custom
  type from and to string, so you can use the enum type directly in your
  struct.
- [objectimpl](../internal/cmd/objectimpl)
  for generating `manifest.Object` implementation for all object kinds.
- [docgen](../internal/cmd/docgen/)
  for generating documentation based on validation rules, Go doc comments and
  generate examples.
- [examplegen](../internal/cmd/examplegen/)
  for generating examples for each manifest object.

Do not edit files marked with `Code generated ... DO NOT EDIT`.
Change the source type, generator input, or generator implementation instead,
then run the relevant generation target.

The mock example under [docs/mock_example](./mock_example) is part of the
workspace and test surface.
Keep it in sync with SDK interface changes.

## Validation

We're using [govy](https://github.com/nobl9/govy) library for validation.
If you encounter any bugs or shortcomings feel free to open an issue or PR
at govy's GitHub page.

## Dependencies

Renovate is configured to automatically merge minor and patch updates.
For major versions, which sadly includes GitHub Actions, manual approval
is required.
