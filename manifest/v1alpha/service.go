package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type ServicesSlice []Service

func (services ServicesSlice) Clone() ServicesSlice {
	clone := make([]Service, len(services))
	copy(clone, services)
	return clone
}

// Service struct which mapped one to one with kind: service yaml definition
type Service struct {
	manifest.ObjectHeader
	Spec   ServiceSpec    `json:"spec"`
	Status *ServiceStatus `json:"status,omitempty"`
}

func (s *Service) GetAPIVersion() string {
	return s.APIVersion
}

func (s *Service) GetKind() manifest.Kind {
	return s.Kind
}

func (s *Service) GetName() string {
	return s.Metadata.Name
}

func (s *Service) Validate() error {
	//TODO implement me
	panic("implement me")
}

func (s *Service) GetProject() string {
	return s.Metadata.Project
}

func (s *Service) SetProject(project string) {
	s.Metadata.Project = project
}

// ServiceStatus represents content of Status optional for Service Object.
type ServiceStatus struct {
	SloCount int `json:"sloCount"`
}

// ServiceSpec represents content of Spec typical for Service Object.
type ServiceSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}

// genericToService converts ObjectGeneric to Object Service.
func genericToService(o manifest.ObjectGeneric, v validator, onlyHeader bool) (Service, error) {
	res := Service{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ServiceSpec
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

// ServiceWithSLOs struct which mapped one to one with kind: service and slo yaml definition.
type ServiceWithSLOs struct {
	Service Service `json:"service"`
	SLOs    []SLO   `json:"slos"`
}
