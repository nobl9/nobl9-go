package nobl9

import "encoding/json"

// Project struct which mapped one to one with kind: project yaml definition.
type Project struct {
	ObjectInternal
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   ProjectMetadata `json:"metadata"`
	Spec       ProjectSpec     `json:"spec"`
}

type ProjectMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

// ProjectSpec represents content of Spec typical for Project Object.
type ProjectSpec struct {
	Description string `json:"description"`
}

// genericToProject converts ObjectGeneric to Project
func genericToProject(o ObjectGeneric, onlyHeader bool) (Project, error) {
	res := Project{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: ProjectMetadata{
			Name:        o.Metadata.Name,
			DisplayName: o.Metadata.DisplayName,
		},
		ObjectInternal: ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ProjectSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec

	return res, nil
}
