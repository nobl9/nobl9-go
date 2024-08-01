# examplegen

`examplegen` is a tool that generates examples for manifest objects.

These examples are usually stored in the `examples` directory of
the respective object's package or in the object's package directory itself
inside `examples.yaml` file.

Each object can have its variants and sub-variants.
Example of such distinction for SLO is Prometheus based SLO with
_calendar aligned_ time window.
Prometheus will be the variant and _calendar aligned_ time window
will be the sub-variant.

## Usage

Run from the repository root's Makefile:

```shell
make generate/examples
```
