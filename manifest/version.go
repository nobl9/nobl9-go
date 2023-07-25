package manifest

//go:generate ../bin/go-enum --names --values --marshal

// Version represents the specific version of the manifest.
// ENUM(v1alpha = n9/v1alpha)
type Version string
