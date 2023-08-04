// Code generated by "generate-object-impl RoleBinding"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ Object = RoleBinding{}

func (r RoleBinding) GetVersion() string {
	return r.APIVersion
}

func (r RoleBinding) GetKind() manifest.Kind {
	return r.Kind
}

func (r RoleBinding) GetName() string {
	return r.Metadata.Name
}

func (r RoleBinding) Validate() error {
	return validator.Check(r)
}

func (r RoleBinding) GetOrganization() string {
	return r.Organization
}

func (r RoleBinding) SetOrganization(org string) manifest.Object {
	r.Organization = org
	return r
}

func (r RoleBinding) GetManifestSource() string {
	return r.ManifestSource
}

func (r RoleBinding) SetManifestSource(src string) manifest.Object {
	r.ManifestSource = src
	return r
}
