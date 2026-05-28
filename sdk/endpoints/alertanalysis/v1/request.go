package v1

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
)

type StartAnalysisRequest struct {
	SLO         string                  `json:"slo"`
	Project     string                  `json:"project"`
	Objective   string                  `json:"objective"`
	StartTime   time.Time               `json:"startTime"`
	EndTime     time.Time               `json:"endTime"`
	AlertPolicy alertpolicy.AlertPolicy `json:"alertPolicy"`
}

type GetAnalysisRequest struct {
	AnalysisID        string
	From              *time.Time
	To                *time.Time
	IncludeTimeseries *bool
}
