package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type ProjectsSlice []Project

func (projects ProjectsSlice) Clone() ProjectsSlice {
	clone := make([]Project, len(projects))
	copy(clone, projects)
	return clone
}

// Project struct which mapped one to one with kind: project yaml definition.
type Project struct {
	manifest.ObjectInternal
	APIVersion string          `json:"apiVersion" validate:"required" example:"n9/v1alpha"`
	Kind       manifest.Kind   `json:"kind" validate:"required" example:"kind"`
	Metadata   ProjectMetadata `json:"metadata"`
	Spec       ProjectSpec     `json:"spec"`
}

func (p Project) GetAPIVersion() string {
	return p.APIVersion
}

func (p Project) GetKind() manifest.Kind {
	return p.Kind
}

func (p Project) GetName() string {
	return p.Metadata.Name
}

func (p Project) Validate() error {
	//TODO implement me
	panic("implement me")
}

type ProjectMetadata struct {
	Name        string          `json:"name" validate:"required,objectName" example:"name"`
	DisplayName string          `json:"displayName,omitempty" validate:"omitempty,min=0,max=63" example:"Shopping App"`
	Labels      manifest.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// ProjectSpec represents content of Spec typical for Project Object.
type ProjectSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}

// genericToProject converts ObjectGeneric to Project
func genericToProject(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Project, error) {
	res := Project{
		APIVersion: o.ObjectHeader.APIVersion,
		Kind:       o.ObjectHeader.Kind,
		Metadata: ProjectMetadata{
			Name:        o.Metadata.Name,
			DisplayName: o.Metadata.DisplayName,
			Labels:      o.Metadata.Labels,
		},
		ObjectInternal: manifest.ObjectInternal{
			Organization: o.ObjectHeader.Organization,
			ManifestSrc:  o.ObjectHeader.ManifestSrc,
		},
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ProjectSpec
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
