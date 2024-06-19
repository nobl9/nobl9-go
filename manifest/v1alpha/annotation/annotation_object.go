// Code generated by "objectimpl Annotation"; DO NOT EDIT.

package annotation

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = Annotation{}
var _ manifest.ProjectScopedObject = Annotation{}
var _ v1alpha.ObjectContext = Annotation{}

func (a Annotation) GetVersion() manifest.Version {
	return a.APIVersion
}

func (a Annotation) GetKind() manifest.Kind {
	return a.Kind
}

func (a Annotation) GetName() string {
	return a.Metadata.Name
}

func (a Annotation) Validate() error {
	if err := validate(a); err != nil {
		return err
	}
	return nil
}

func (a Annotation) GetManifestSource() string {
	return a.ManifestSource
}

func (a Annotation) SetManifestSource(src string) manifest.Object {
	a.ManifestSource = src
	return a
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

func (a Annotation) GetValidator() validation.Validator[Annotation] {
	return validator
}
