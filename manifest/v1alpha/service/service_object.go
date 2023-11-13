// Code generated by "generate-object-impl Service"; DO NOT EDIT.

package service

import "github.com/nobl9/nobl9-go/manifest"

// Ensure interfaces are implemented.
var _ manifest.Object = Service{}
var _ manifest.ProjectScopedObject = Service{}

func (s Service) GetVersion() string {
	return s.APIVersion
}

func (s Service) GetKind() manifest.Kind {
	return s.Kind
}

func (s Service) GetName() string {
	return s.Metadata.Name
}

func (s Service) Validate() error {
	if err := validate(s); err != nil {
		return err
	}
	return nil
}

func (s Service) GetProject() string {
	return s.Metadata.Project
}

func (s Service) SetProject(project string) manifest.Object {
	s.Metadata.Project = project
	return s
}

func (s Service) GetOrganization() string {
	return s.Organization
}

func (s Service) SetOrganization(org string) manifest.Object {
	s.Organization = org
	return s
}

func (s Service) GetManifestSource() string {
	return s.ManifestSource
}

func (s Service) SetManifestSource(src string) manifest.Object {
	s.ManifestSource = src
	return s
}
