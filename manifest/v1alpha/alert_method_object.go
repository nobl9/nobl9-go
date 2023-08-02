// Code generated by "generate-object-impl AlertMethod"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

func (a AlertMethod) GetVersion() string {
	return a.APIVersion
}

func (a AlertMethod) GetKind() manifest.Kind {
	return a.Kind
}

func (a AlertMethod) GetName() string {
	return a.Metadata.Name
}

func (a AlertMethod) Validate() error {
	return validator.Check(a)
}

func (a AlertMethod) GetProject() string {
	return a.Metadata.Project
}

func (a AlertMethod) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a AlertMethod) GetOrganization() string {
	return a.Organization
}

func (a AlertMethod) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}

func (a AlertMethod) GetManifestSource() string {
	return a.ManifestSource
}

func (a AlertMethod) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}
