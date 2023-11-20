// Code generated by "generate-object-impl Project"; DO NOT EDIT.

package project

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = Project{}
var _ v1alpha.ObjectContext = Project{}

func (p Project) GetVersion() string {
	return p.APIVersion
}

func (p Project) GetKind() manifest.Kind {
	return p.Kind
}

func (p Project) GetName() string {
	return p.Metadata.Name
}

func (p Project) Validate() error {
	if err := validate(p); err != nil {
		return err
	}
	return nil
}

func (p Project) GetManifestSource() string {
	return p.ManifestSource
}

func (p Project) SetManifestSource(src string) manifest.Object {
	p.ManifestSource = src
	return p
}

func (p Project) GetOrganization() string {
	return p.Organization
}

func (p Project) SetOrganization(org string) manifest.Object {
	p.Organization = org
	return p
}
