# docgen

`docgen` is a tool that generates documentation for manifest objects.
It merges:

- Go doc comments
- [Validation plan](../../validation/plan.go)
- Custom documentation and formatting
- Generated examples

into a single YAML file.

## Usage

```shell
go run . -o docs.yaml
```
