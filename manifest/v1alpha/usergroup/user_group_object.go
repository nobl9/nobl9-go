// Code generated by "objectimpl UserGroup"; DO NOT EDIT.

package usergroup

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = UserGroup{}
var _ v1alpha.ObjectContext = UserGroup{}

func (u UserGroup) GetVersion() manifest.Version {
	return u.APIVersion
}

func (u UserGroup) GetKind() manifest.Kind {
	return u.Kind
}

func (u UserGroup) GetName() string {
	return u.Metadata.Name
}

func (u UserGroup) Validate() error {
	if err := validate(u); err != nil {
		return err
	}
	return nil
}

func (u UserGroup) GetManifestSource() string {
	return u.ManifestSource
}

func (u UserGroup) SetManifestSource(src string) manifest.Object {
	u.ManifestSource = src
	return u
}

func (u UserGroup) GetOrganization() string {
	return u.Organization
}

func (u UserGroup) SetOrganization(org string) manifest.Object {
	u.Organization = org
	return u
}

func (u UserGroup) GetValidator() govy.Validator[UserGroup] {
	return validator
}
