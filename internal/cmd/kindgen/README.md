# kindgen

`kindgen` generates the `ProjectScopedKinds()` function in the `manifest` package
by scanning all `*_object.go` files in `manifest/v1alpha/` for types that implement
`manifest.ProjectScopedObject`.

## How it works

1. Walks through `manifest/v1alpha/**/*_object.go` files.
2. Parses each file looking for `var _ manifest.ProjectScopedObject = TypeName{}`.
3. Extracts type names and maps them to `Kind` constants (e.g., `SLO` -> `KindSLO`).
4. Generates `manifest/kind_project_scoped.go` with the `ProjectScopedKinds()` function.

## Usage

This generator is invoked automatically by `make generate/code` after `go generate`
completes. This ordering ensures that `objectimpl` has already created the
`*_object.go` files before `kindgen` scans them.

```bash
make generate/code
```

To run manually (from the repository root):

```bash
cd manifest && go run ../internal/cmd/kindgen
```
