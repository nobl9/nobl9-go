package nobl9

import (
	"encoding/json"
)

type UserGroup struct {
	ObjectInternal
	APIVersion string        `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       string        `json:"kind" validate:"required" example:"kind"`
	Metadata   GroupMetadata `json:"metadata"`
	Spec       UserGroupSpec `json:"spec"`
}

// UserGroupSpec represents content of UserGroup's Spec
type UserGroupSpec struct {
	DisplayName string   `json:"displayName"`
	Members     []Member `json:"members"`
}

type Member struct {
	ID string `json:"id"`
}

type GroupMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}

// genericToUserGroup converts ObjectGeneric to UserGroup object
func genericToUserGroup(o ObjectGeneric) (UserGroup, error) {
	res := UserGroup{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: GroupMetadata{
			Name: o.Metadata.Name,
		},
		ObjectInternal: ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	var resSpec UserGroupSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec

	return res, nil
}
