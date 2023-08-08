// Code generated by "generate-object-impl AlertPolicy"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = AlertPolicy{}
var _ manifest.ProjectScopedObject = AlertPolicy{}
var _ ObjectContext = AlertPolicy{}

func (a AlertPolicy) GetVersion() string {
	return a.APIVersion
}

func (a AlertPolicy) GetKind() manifest.Kind {
	return a.Kind
}

func (a AlertPolicy) GetName() string {
	return a.Metadata.Name
}

func (a AlertPolicy) Validate() error {
	return validator.Check(a)
}

func (a AlertPolicy) GetProject() string {
	return a.Metadata.Project
}

func (a AlertPolicy) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a AlertPolicy) GetOrganization() string {
	return a.Organization
}

func (a AlertPolicy) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}

func (a AlertPolicy) GetManifestSource() string {
	return a.ManifestSource
}

func (a AlertPolicy) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}
