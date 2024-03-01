# Development

This document describes the intricacies of nobl9-go development workflow.
If you see anything missing, feel free to contribute :)

## Pull requests

[Pull request template](.github/pull_request_template.md)
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

## Testing

Currently nobl9-go is automatically verified with unit tests only.
It is encouraged to create a simple MVP program which verifies the introduced
changes work. There's a dedicated section in PR template `## Testing` which
is a great place to add such `main.go` code snippet.
Here's an example:

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/nobl9/nobl9-go/sdk"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func main() {
	client, err := sdk.DefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	projects, err := client.Objects().V1().GetV1alphaProjects(context.Background(), v1.GetProjectsRequest{
		Names: []string{"default"},
	})
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(projects)
}
```

### Unit tests

When writing validation for Nobl9 manifest objects, adhere to the following
rules:

- Use `validation` package ([see](#validation)).
- **ALWAYS** test the whole object and not only its specific fields.
  *TIP*: Create "valid" object once and then just modify its specific fields
  to validate them.
- **ALWAYS** use `testutils` package and its `AssertNoError` and
  `AssertContainsErrors`. It not only makes it easier to validate the whole
  object but also it allows recording these tests.
  Recorded tests are planned to be used for regression and dependent
  tools (sloctl, Terraform provider) testing.

### Recording tests

If you wish to record the tests run `make test/record`.
By default, the tests are recorded inside `./bin` folder.

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
- [Our custom tool](scripts/generate-object-impl.go)
  for generating `manifest.Object` methods implementation for all object kinds.

## Validation

We're using our own validation library to write validation for all objects.
Refer to this [README.md](../internal/validation/README.md) for more information.

## Dependencies

Renovate is configured to automatically merge minor and patch updates.
For major versions, which sadly includes GitHub Actions, manual approval
is required.
