package models

// MoveSLOs is a payload for move SLOs API which allows moving SLOs between projects.
type MoveSLOs struct {
	// SLONames is a list of SLO names to move between projects.
	SLONames []string `json:"sloNames"`
	// OldProject is the current project name of the moved SLOs.
	OldProject string `json:"oldProject"`
	// NewProject is the project name to which the SLOs is moved.
	NewProject string `json:"newProject"`
	// Service is the target service name to which the moved SLOs is assigned.
	Service string `json:"service"`
	// DetachAlertPolicies defines If the moved SLOs should have their alert policies automatically detached.
	// It defaults to false.
	DetachAlertPolicies bool `json:"detachAlertPolicies"`
}
