// Code generated by "objectimpl Report"; DO NOT EDIT.

package report

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = Report{}
var _ v1alpha.ObjectContext = Report{}

func (r Report) GetVersion() manifest.Version {
	return r.APIVersion
}

func (r Report) GetKind() manifest.Kind {
	return r.Kind
}

func (r Report) GetName() string {
	return r.Metadata.Name
}

func (r Report) Validate() error {
	if err := validate(r); err != nil {
		return err
	}
	return nil
}

func (r Report) GetManifestSource() string {
	return r.ManifestSource
}

func (r Report) SetManifestSource(src string) manifest.Object {
	r.ManifestSource = src
	return r
}

func (r Report) GetOrganization() string {
	return r.Organization
}

func (r Report) SetOrganization(org string) manifest.Object {
	r.Organization = org
	return r
}

func (r Report) GetValidator() govy.Validator[Report] {
	return validator
}
