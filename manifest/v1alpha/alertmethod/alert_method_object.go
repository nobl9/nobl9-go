// Code generated by "objectimpl AlertMethod"; DO NOT EDIT.

package alertmethod

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = AlertMethod{}
var _ manifest.ProjectScopedObject = AlertMethod{}
var _ v1alpha.ObjectContext = AlertMethod{}

func (a AlertMethod) GetVersion() manifest.Version {
	return a.APIVersion
}

func (a AlertMethod) GetKind() manifest.Kind {
	return a.Kind
}

func (a AlertMethod) GetName() string {
	return a.Metadata.Name
}

func (a AlertMethod) Validate() error {
	if err := validate(a); err != nil {
		return err
	}
	return nil
}

func (a AlertMethod) GetManifestSource() string {
	return a.ManifestSource
}

func (a AlertMethod) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
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

func (a AlertMethod) GetValidator() govy.Validator[AlertMethod] {
	return validator
}
