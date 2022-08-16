package nobl9

import "encoding/json"

// RoleBinding represents relation of User and Role
type RoleBinding struct {
	ObjectInternal
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   RoleBindingMetadata `json:"metadata"`
	Spec       RoleBindingSpec     `json:"spec"`
}

type RoleBindingSpec struct {
	User       string `json:"user"`
	RoleRef    string `json:"roleRef"`
	ProjectRef string `json:"projectRef,omitempty"`
}

type RoleBindingMetadata struct {
	Name string `json:"name"`
}

// genericToRoleBinding converts ObjectGeneric to ObjectRoleBinding
// onlyHeader parameter is not supported for RoleBinding since ProjectRef is defined on Spec section.
func genericToRoleBinding(o ObjectGeneric) (RoleBinding, error) {
	res := RoleBinding{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: RoleBindingMetadata{
			Name: o.Metadata.Name,
		},
		ObjectInternal: ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	var resSpec RoleBindingSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec

	return res, nil
}
