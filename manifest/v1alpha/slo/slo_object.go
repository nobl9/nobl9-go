// Code generated by "generate-object-impl SLO"; DO NOT EDIT.

package slo

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = SLO{}
var _ manifest.ProjectScopedObject = SLO{}
var _ v1alpha.ObjectContext = SLO{}

func (s SLO) GetVersion() manifest.Version {
	return s.APIVersion
}

func (s SLO) GetKind() manifest.Kind {
	return s.Kind
}

func (s SLO) GetName() string {
	return s.Metadata.Name
}

func (s SLO) Validate() error {
	if err := validate(s); err != nil {
		return err
	}
	return nil
}

func (s SLO) GetManifestSource() string {
	return s.ManifestSource
}

func (s SLO) SetManifestSource(src string) manifest.Object {
	s.ManifestSource = src
	return s
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

func (s SLO) GetValidator() validation.Validator[SLO] {
	return validator
}
