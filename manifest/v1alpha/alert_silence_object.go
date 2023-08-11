// Code generated by "generate-object-impl AlertSilence"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = AlertSilence{}
var _ manifest.ProjectScopedObject = AlertSilence{}
var _ ObjectContext = AlertSilence{}

func (a AlertSilence) GetVersion() string {
	return a.APIVersion
}

func (a AlertSilence) GetKind() manifest.Kind {
	return a.Kind
}

func (a AlertSilence) GetName() string {
	return a.Metadata.Name
}

func (a AlertSilence) Validate() error {
	return validator.Check(a)
}

func (a AlertSilence) GetProject() string {
	return a.Metadata.Project
}

func (a AlertSilence) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a AlertSilence) GetOrganization() string {
	return a.Organization
}

func (a AlertSilence) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}

func (a AlertSilence) GetManifestSource() string {
	return a.ManifestSource
}

func (a AlertSilence) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}
