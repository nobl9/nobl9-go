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

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (s Service) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: s.Metadata.Name, Project: s.Metadata.Project}
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
