package slo

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

// CompositeSpec represents a composite of SLOs and Composite SLOs.
type CompositeSpec struct {
	MaxDelay   v1alpha.Duration `json:"maxDelay"`
	Components `json:"components"`
}

type Components struct {
	Objectives []CompositeObjective `json:"objectives"`
}

type CompositeObjective struct {
	Project     string   `json:"project"`
	SLO         string   `json:"slo"`
	Objective   string   `json:"objective"`
	Weight      *float64 `json:"weight,omitempty"`
	WhenDelayed string   `json:"whenDelayed"`
}

// WhenDelayedEnum represents enum for behaviour of Composite SLO objectives
type WhenDelayedEnum int16

const (
	CountAsGood WhenDelayedEnum = iota + 1
	CountAsBad
	Ignore
)

func (s WhenDelayedEnum) String() string {
	switch s {
	case CountAsGood:
		return "CountAsGood"
	case CountAsBad:
		return "CountAsBad"
	case Ignore:
		return "Ignore"
	}
	return ""
}
