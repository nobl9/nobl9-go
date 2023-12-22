<!-- markdownlint-disable line-length html -->
<h1 align="center">
   <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/32738712/185149468-dc07f5d9-68c0-4922-a006-7baf6a08eaac.png">
      <source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/32738712/185148352-bea80385-c772-4842-8f7b-6838bb08a3f4.png">
      <img alt="N9" src="https://user-images.githubusercontent.com/32738712/185148352-bea80385-c772-4842-8f7b-6838bb08a3f4.png" width="500" />
   </picture><br/>
</h1>

<div align="center">
  <table>
    <tr>
      <td>
        <img alt="checks" src="https://github.com/nobl9/nobl9-go/actions/workflows/checks.yml/badge.svg?event=push">
      </td>
      <td>
        <img alt="tests" src="https://github.com/nobl9/nobl9-go/actions/workflows/tests.yml/badge.svg?event=push">
      </td>
      <td>
        <img alt="vulnerabilities" src="https://github.com/nobl9/nobl9-go/actions/workflows/vulns.yml/badge.svg?event=push">
      </td>
    </tr>
  </table>
</div>
<!-- markdownlint-enable line-length html -->

Nobl9 SDK for the Go programming language.

Checkout [release notes](https://github.com/nobl9/nobl9-go/releases)
for details on the latest bug fixes, updates, and features.

⚠️ Until v1.0.0 is released we expect some minor breaking API changes
to occur.

---

Legend:

1. [Installation](#installation)
2. [Examples](#examples)
3. [Repository structure](#repository-structure)
4. [Contributing](#contributing)

# Installation

To add the latest version to your Go module run:

```shell
go get github.com/nobl9/nobl9-go
```

# Examples

## Basic usage

<!-- markdownlint-disable MD013 -->
```go
package main

import (
   "context"
   "encoding/json"
   "fmt"
   "log"

   "github.com/nobl9/nobl9-go/manifest"
   "github.com/nobl9/nobl9-go/manifest/v1alpha"
   "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
   "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
   "github.com/nobl9/nobl9-go/sdk"
   objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func main() {
   ctx := context.Background()

   // Create client.
   client, err := sdk.DefaultClient()
   if err != nil {
      log.Fatalf("failed to create sdk client, err: %v", err)
   }

   // Read from file, url or glob pattern.
   objects, err := sdk.ReadObjects(ctx, "./project.yaml")
   if err != nil {
      log.Fatalf("failed to read project.yaml file, err: %v", err)
   }
   // Use manifest.FilterByKind to extract specific objects from the manifest.Object slice.
   myProject := manifest.FilterByKind[project.Project](objects)[0]
   // Define objects in code.
   myService := service.New(
      service.Metadata{
         Name:        "my-service",
         DisplayName: "My Service",
         Project:     myProject.GetName(),
         Labels: v1alpha.Labels{
            "team":   []string{"green", "orange"},
            "region": []string{"eu-central-1"},
         },
      },
      service.Spec{
         Description: "Example service",
      },
   )
   objects = append(objects, myService)

   // Verify the objects.
   if errs := manifest.Validate(objects); len(errs) > 0 {
      log.Fatalf("service validation failed, errors: %v", errs)
   }

   // Apply the objects.
   if err = client.Objects().V1().Apply(ctx, objects); err != nil {
      log.Fatalf("failed to apply objects, err: %v", err)
   }

   // Get the applied resources.
   services, err := client.Objects().V1().GetV1alphaServices(ctx, objectsV1.GetServicesRequest{
      Project: myProject.GetName(),
      Names:   []string{myService.GetName()},
   })
   if err != nil {
      log.Fatalf("failed to get services, err: %v", err)
   }
   projects, err := client.Objects().V1().GetV1alphaProjects(ctx, objectsV1.GetProjectsRequest{
      Names: []string{myProject.GetName()},
   })
   if err != nil {
      log.Fatalf("failed to get projects, err: %v", err)
   }

   // Aggregate objects back into manifest.Objects slice.
   appliedObjects := make([]manifest.Object, 0, len(services)+len(projects))
   for _, service := range services {
      appliedObjects = append(appliedObjects, service)
   }
   for _, project := range projects {
      appliedObjects = append(appliedObjects, project)
   }

   // Print JSON representation of these objects.
   data, err := json.MarshalIndent(appliedObjects, "", "  ")
   if err != nil {
      log.Fatalf("failed to marshal objects, err: %v", err)
   }
   fmt.Println(string(data))

   // Delete resources.
   if err = client.Objects().V1().Delete(ctx, objects); err != nil {
      log.Fatalf("failed to delete objects, err: %v", err)
   }
}
```
<!-- markdownlint-enable MD013 -->

# Repository structure

## Public packages

1. [sdk](./sdk) defines:

    - `Client` which exposes methods for interacting with
      different Nobl9 web APIs.
    - Methods for reading and managing Nobl9
      configuration (including `config.toml` file) used by tools
      such as `sloctl` or the SDK itself.
    - Methods for fetching and parsing Nobl9 configuration objects.

2. [manifest](./manifest) holds definitions of all Nobl9 configuration
   objects, such as SLO or Project. It is divided into three package
   levels:

    - [manifest](./manifest) defines general contracts and generic methods
      for all objects.
    - Version specific packages, like [v1alpha](./manifest/v1alpha), define
      version specific API shared by multiple objects.
    - Object specific packages, like [slo](./manifest/v1alpha/slo), provide
      object definition for specific object version.

   ```text
   └── manifest
       └── version (e.g. v1alpha)
           └── object (e.g. slo)
   ```

# Contributing

TBA
