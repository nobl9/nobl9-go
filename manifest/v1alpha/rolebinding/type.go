package rolebinding

//go:generate ../../../bin/go-enum  --nocase --lower --names

// Type represents the type of the [RoleBinding].
// ENUM(Project = 1, Organization)
type Type int
