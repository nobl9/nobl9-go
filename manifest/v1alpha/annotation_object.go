// Code generated by "generate-object-impl Annotation"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = Annotation{}
var _ manifest.ProjectScopedObject = Annotation{}
var _ ObjectContext = Annotation{}

func (a Annotation) GetVersion() string {
	return a.APIVersion
}

func (a Annotation) GetKind() manifest.Kind {
	return a.Kind
}

func (a Annotation) GetName() string {
	return a.Metadata.Name
}

func (a Annotation) Validate() error {
	return validator.Check(a)
}

func (a Annotation) GetProject() string {
	return a.Metadata.Project
}

func (a Annotation) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a Annotation) GetOrganization() string {
	return a.Organization
}

func (a Annotation) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}

func (a Annotation) GetManifestSource() string {
	return a.ManifestSource
}

func (a Annotation) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}
