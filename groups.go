package nobl9

import (
	"encoding/json"
)

type Group struct {
	ObjectInternal
	APIVersion string        `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       string        `json:"kind" validate:"required" example:"kind"`
	Metadata   GroupMetadata `json:"metadata"`
	Spec       GroupSpec     `json:"spec"`
}

// GroupSpec represents content of Group's Spec
type GroupSpec struct {
	DisplayName string   `json:"displayName"`
	Members     []Member `json:"members"`
}

type Member struct {
	ID string `json:"id"`
}

type GroupMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}

// genericToGroup converts ObjectGeneric to Group object
func genericToGroup(o ObjectGeneric) (Group, error) {
	res := Group{
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
	var resSpec GroupSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec

	return res, nil
}
