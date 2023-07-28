package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type UserGroupsSlice []UserGroup

func (u UserGroupsSlice) Clone() UserGroupsSlice {
	clone := make([]UserGroup, len(u))
	copy(clone, u)
	return clone
}

type UserGroup struct {
	manifest.ObjectInternal
	APIVersion string        `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       manifest.Kind `json:"kind" validate:"required" example:"kind"`
	Metadata   GroupMetadata `json:"metadata"`
	Spec       UserGroupSpec `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (u UserGroup) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: u.Metadata.Name}
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
func genericToUserGroup(o manifest.ObjectGeneric) (UserGroup, error) {
	res := UserGroup{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: GroupMetadata{
			Name: o.Metadata.Name,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	var resSpec UserGroupSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec

	return res, nil
}
