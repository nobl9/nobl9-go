# nobl9-go
Nobl9 API client in Go, primarily used by the [Nobl9](https://github.com/nobl9/terraform-provider-nobl9) provider in Terraform.


## Installation
```bash
go get github.com/nobl9/nobl9-go
```

## Example usage
```go
package main

import (
    "fmt"

    "github.com/nobl9/nobl9-go"
)

func main() {
        o := nobl9.Service{
        ObjectHeader: nobl9.ObjectHeader{
            APIVersion: "n9/v1alpha",
            Kind:       "Service",
            MetadataHolder: nobl9.MetadataHolder{
                Metadata: nobl9.Metadata{
                    Name:        "sdk-test",
                    DisplayName: "SDK test",
                    Project:     "sdk-test",
                },
            },
        },
        Spec: nobl9.ServiceSpec{
            Description: "Test service made by SDK",
        },
    }

    var p nobl9.Payload
    p.AddObject(o)
    c, _ := nobl9.NewClient("https://main.nobl9.dev/api",
        "nobl9-dev", "test", "nobl9-go",
        "[CLIENT_ID]", "CLIENT_SECRET",
        "https://accounts.nobl9.dev", "auseg9kiegWKEtJZC416")   

    err := c.ApplyObjects(p.GetObjects())
    fmt.Println(err)
}
```

## Contributing
