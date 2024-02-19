package slo

// CompositeSpec represents a composite of SLOs and Composite SLOs.
type CompositeSpec struct {
	MaxDelay   string `json:"maxDelay"`
	Components `json:"components"`
}

type Components struct {
	Objectives []CompositeObjective `json:"objectives"`
}

type CompositeObjective struct {
	Project     string  `json:"project"`
	SLO         string  `json:"slo"`
	Objective   string  `json:"objective"`
	Weight      float64 `json:"weight"`
	WhenDelayed string  `json:"whenDelayed"`
}

// WhenDelayedEnum represents enum for behavior of Composite SLO objectives
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
