// Code generated by "generate-object-impl Agent"; DO NOT EDIT.

package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = Agent{}
var _ manifest.ProjectScopedObject = Agent{}
var _ ObjectContext = Agent{}

func (a Agent) GetVersion() string {
	return a.APIVersion
}

func (a Agent) GetKind() manifest.Kind {
	return a.Kind
}

func (a Agent) GetName() string {
	return a.Metadata.Name
}

func (a Agent) Validate() error {
	return validator.Check(a)
}

func (a Agent) GetManifestSource() string {
	return a.ManifestSource
}

func (a Agent) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}

func (a Agent) GetProject() string {
	return a.Metadata.Project
}

func (a Agent) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a Agent) GetOrganization() string {
	return a.Organization
}

func (a Agent) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}
