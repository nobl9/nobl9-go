# docgen

`objectimpl` is a tool that generates boilerplate functions for manifest objects
in order to implement `manifest.Object` interface.

It utilizes `text/template` to generate the code.

## Usage

Add the following `generate` directive to the file that contains
the `manifest.Object` object definition.
Replace the `<OBJECT_KIND>` with the struct name of your object, e.g. `Project`.

```go
//go:generate go run ../../../internal/cmd/objectimpl <OBJECT_KIND>
```
