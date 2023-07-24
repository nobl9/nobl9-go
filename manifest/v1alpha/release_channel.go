package v1alpha

//go:generate ../../bin/go-enum --nocase --lower --names --values --marshal

// ReleaseChannel /* ENUM(Stable,Beta,Alpha)*/
type ReleaseChannel int
