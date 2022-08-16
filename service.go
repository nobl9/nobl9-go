package nobl9

import "encoding/json"

// Service struct which mapped one to one with kind: service yaml definition
type Service struct {
	ObjectHeader
	Spec   ServiceSpec   `json:"spec"`
	Status ServiceStatus `json:"status"`
}

// ServiceWithSLOs struct which mapped one to one with kind: service and slo yaml definition
type ServiceWithSLOs struct {
	Service Service `json:"service"`
	SLOs    []SLO   `json:"slos"`
}

// ServiceStatus represents content of Status optional for Service Object
type ServiceStatus struct {
	SloCount int `json:"sloCount"`
}

// ServiceSpec represents content of Spec typical for Service Object
type ServiceSpec struct {
	Description string `json:"description"`
}

// genericToService converts ObjectGeneric to Object Service
func genericToService(o ObjectGeneric, onlyHeader bool) (Service, error) {
	res := Service{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}

	var resSpec ServiceSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	return res, nil
}
