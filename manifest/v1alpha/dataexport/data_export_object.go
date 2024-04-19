// Code generated by "generate-object-impl DataExport"; DO NOT EDIT.

package dataexport

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = DataExport{}
var _ manifest.ProjectScopedObject = DataExport{}
var _ v1alpha.ObjectContext = DataExport{}

func (d DataExport) GetVersion() manifest.Version {
	return d.APIVersion
}

func (d DataExport) GetKind() manifest.Kind {
	return d.Kind
}

func (d DataExport) GetName() string {
	return d.Metadata.Name
}

func (d DataExport) Validate() error {
	if err := validate(d); err != nil {
		return err
	}
	return nil
}

func (d DataExport) GetManifestSource() string {
	return d.ManifestSource
}

func (d DataExport) SetManifestSource(src string) manifest.Object {
	d.ManifestSource = src
	return d
}

func (d DataExport) GetProject() string {
	return d.Metadata.Project
}

func (d DataExport) SetProject(project string) manifest.Object {
	d.Metadata.Project = project
	return d
}

func (d DataExport) GetOrganization() string {
	return d.Organization
}

func (d DataExport) SetOrganization(org string) manifest.Object {
	d.Organization = org
	return d
}

func (d DataExport) GetValidator() validation.Validator[DataExport] {
	return validator
}
