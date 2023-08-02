package manifest

//go:generate ../bin/go-enum --names

// ObjectFormat represents the format of Object data representation.
// ENUM(JSON = 1, YAML)
type ObjectFormat int
