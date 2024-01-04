## Motivation

Describe what is the motivation behind the proposed changes. If possible reference the current solution/state of affairs.

## Summary

Recap of changed code.

## Related changes

List related changes from other PRs (if any).

## Testing

- Describe how to check introduced code changes manually. Simple `main.go` which takes advantage of the introduced changes is preferred (if possible and useful).
- Take care of test coverage on unit, integration or even end-to-end tests level.

## Checklist

- [ ] Include this change in Release Notes?
  - If yes, write 1-3 sentences about the changes here and explicitly list all changes that can surprise our users.
- [ ] Are these changes required to be in sync with the API? Example of such can be extending a `manifest.Object` with a new field.
      It won't be usable until Nobl9 platform version is rolled out which can handle this field.
  - If yes, **MUST NOT** create an official release, use a pre-release version, like `v1.1.0-rc1` instead.
  - If the changes are independent of Nobl9 platform version, you can release an offical version, like `v1.1.0`.
