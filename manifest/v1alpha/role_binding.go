package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type RoleBindingsSlice []RoleBinding

func (roleBindings RoleBindingsSlice) Clone() RoleBindingsSlice {
	clone := make([]RoleBinding, len(roleBindings))
	copy(clone, roleBindings)
	return clone
}

// RoleBinding represents relation of User and Role
type RoleBinding struct {
	manifest.ObjectInternal
	APIVersion string              `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       manifest.Kind       `json:"kind" validate:"required" example:"kind"`
	Metadata   RoleBindingMetadata `json:"metadata"`
	Spec       RoleBindingSpec     `json:"spec"`
}

func (r RoleBinding) GetAPIVersion() string {
	return r.APIVersion
}

func (r RoleBinding) GetKind() manifest.Kind {
	return r.Kind
}

func (r RoleBinding) GetName() string {
	return r.Metadata.Name
}

func (r RoleBinding) Validate() error {
	//TODO implement me
	panic("implement me")
}

type RoleBindingMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}

type RoleBindingSpec struct {
	User       *string `json:"user,omitempty" validate:"required_without=GroupRef"`
	GroupRef   *string `json:"groupRef,omitempty" validate:"required_without=User"`
	RoleRef    string  `json:"roleRef" validate:"required"`
	ProjectRef string  `json:"projectRef,omitempty"`
}

// genericToRoleBinding converts ObjectGeneric to ObjectRoleBinding
// onlyHeader parameter is not supported for RoleBinding since ProjectRef is defined on Spec section.
func genericToRoleBinding(o manifest.ObjectGeneric, v validator) (RoleBinding, error) {
	res := RoleBinding{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: RoleBindingMetadata{
			Name: o.Metadata.Name,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	var resSpec RoleBindingSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	return res, nil
}
