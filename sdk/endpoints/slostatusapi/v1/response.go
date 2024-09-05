package v1

import "github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi"

type SLOListResponse struct {
	Data  []SLODetails       `json:"data"`
	Links slostatusapi.Links `json:"links"`
}

type SLODetails struct {
	Name            string                   `json:"name"`
	DisplayName     string                   `json:"displayName"`
	Description     string                   `json:"description"`
	Project         string                   `json:"project"`
	Service         string                   `json:"service"`
	CreatedAt       string                   `json:"createdAt"`
	Objectives      []slostatusapi.Objective `json:"objectives"`
	Composite       *slostatusapi.Composite  `json:"composite,omitempty"`
	Labels          map[string][]string      `json:"labels,omitempty"`
	Annotations     map[string]string        `json:"annotations,omitempty"`
	BudgetingMethod string                   `json:"budgetingMethod"`
}
