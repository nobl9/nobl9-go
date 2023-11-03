// Code generated by "generate-object-impl SLO"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = SLO{}
var _ manifest.ProjectScopedObject = SLO{}
var _ ObjectContext = SLO{}

func (s SLO) GetVersion() string {
	return s.APIVersion
}

func (s SLO) GetKind() manifest.Kind {
	return s.Kind
}

func (s SLO) GetName() string {
	return s.Metadata.Name
}

func (s SLO) Validate() error {
	return validator.Check(s)
}

func (s SLO) GetProject() string {
	return s.Metadata.Project
}

func (s SLO) SetProject(project string) manifest.Object {
	s.Metadata.Project = project
	return s
}

func (s SLO) GetOrganization() string {
	return s.Organization
}

func (s SLO) SetOrganization(org string) manifest.Object {
	s.Organization = org
	return s
}

func (s SLO) GetManifestSource() string {
	return s.ManifestSource
}

func (s SLO) SetManifestSource(src string) manifest.Object {
	s.ManifestSource = src
	return s
}