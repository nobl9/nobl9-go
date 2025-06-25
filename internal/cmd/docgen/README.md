# docgen

`docgen` is a tool that generates documentation for manifest objects.
It merges:

- Go doc comments
- [Validation plan](https://pkg.go.dev/github.com/nobl9/govy@latest/pkg/govy#Plan)
- Custom documentation and formatting
- Generated examples

into a single YAML file.

## Usage

```shell
go run . -o docs.yaml
```
