// Code generated by "generate-object-impl Alert"; DO NOT EDIT.

package alert

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = Alert{}
var _ manifest.ProjectScopedObject = Alert{}
var _ v1alpha.ObjectContext = Alert{}

func (a Alert) GetVersion() manifest.Version {
	return a.APIVersion
}

func (a Alert) GetKind() manifest.Kind {
	return a.Kind
}

func (a Alert) GetName() string {
	return a.Metadata.Name
}

func (a Alert) Validate() error {
	if err := validate(a); err != nil {
		return err
	}
	return nil
}

func (a Alert) GetManifestSource() string {
	return a.ManifestSource
}

func (a Alert) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
}

func (a Alert) GetProject() string {
	return a.Metadata.Project
}

func (a Alert) SetProject(project string) manifest.Object {
	a.Metadata.Project = project
	return a
}

func (a Alert) GetOrganization() string {
	return a.Organization
}

func (a Alert) SetOrganization(org string) manifest.Object {
	a.Organization = org
	return a
}

func (a Alert) GetValidator() validation.Validator[Alert] {
	return validator
}
