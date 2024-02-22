# Development

This document describes the intricacies of nobl9-go development workflow.

## Makefile

Run `make help` to display short description for each target.
The provided Makefile will automatically install dev dependencies if they're
missing and place them under `bin`
(this does not apply to `yarn` managed dependencies).
However, it does not detect if the binary you have is up to date with the
versions declaration located in Makefile.
If you see any discrepancies between CI and your local runs, remove the
binaries from `bin` and let Makefile reinstall them with the latest version.

## Testing

Currently nobl9-go is automatically only verified with unit tests.
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

## Releases

Refer to [RELEASE.md](./RELEASE.md) for more information on release process.

## Code generation

TODO

## Dependencies

Renovate is configured to automatically merge minor and patch updates.
For major versions, which sadly includes GitHub Actions, manual approval
is required.
