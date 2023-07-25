package v1alpha

//go:generate ../../bin/go-enum --nocase --lower --names --values --marshal

// ReleaseChannel /* ENUM(stable = 1, beta, alpha)*/
type ReleaseChannel int
