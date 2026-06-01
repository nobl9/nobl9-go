package v1

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
)

type StartAnalysisResponse struct {
	AnalysisID string `json:"analysisId"`
}

type CalculateAlertPolicyResponse struct {
	AlertPolicies          []alertpolicy.AlertPolicy `json:"alertPolicies"`
	AdjustedStartTime      time.Time                 `json:"adjustedStartTime"`
	AdjustedEndTime        time.Time                 `json:"adjustedEndTime"`
	RemainingBudgetAtStart float64                   `json:"remainingBudgetAtStart"`
	RemainingBudgetAtEnd   float64                   `json:"remainingBudgetAtEnd"`
}

type GetAnalysisResponse struct {
	Alerts           []alert.Alert           `json:"alerts"`
	AlertPolicy      alertpolicy.AlertPolicy `json:"alertPolicy"`
	SLO              string                  `json:"slo"`
	Project          string                  `json:"project"`
	Objective        string                  `json:"objective"`
	StartTime        time.Time               `json:"startTime"`
	EndTime          time.Time               `json:"endTime"`
	Status           Status                  `json:"status"`
	DetectionStatus  ReadinessStatus         `json:"detectionStatus"`
	TimeseriesStatus ReadinessStatus         `json:"timeseriesStatus"`
	Timeseries       []Timeseries            `json:"timeseries"`
}

type Status string

const (
	StatusExportingTimeseries      Status = "exporting_timeseries"
	StatusCalculatingAlertsmetrics Status = "calculating_alertsmetrics"
	StatusCalculatingAlerts        Status = "calculating_alerts"
	StatusStoreAlerts              Status = "store_alerts"
	StatusCreatingVersion          Status = "creating_version"
	StatusDownsampling             Status = "downsampling"
	StatusUpdatingTimeseries       Status = "updating_timeseries"
	StatusEnablingVersion          Status = "enabling_version"
	StatusDone                     Status = "done"
	StatusCanceled                 Status = "canceled"
	StatusError                    Status = "error"
)

type ReadinessStatus string

const (
	ReadinessStatusPending  ReadinessStatus = "pending"
	ReadinessStatusRunning  ReadinessStatus = "running"
	ReadinessStatusReady    ReadinessStatus = "ready"
	ReadinessStatusCanceled ReadinessStatus = "canceled"
	ReadinessStatusError    ReadinessStatus = "error"
)

type Timeseries struct {
	Measurement string         `json:"measurement"`
	Timestamps  []int64        `json:"timestamps"`
	Values      []float64      `json:"values"`
	Attributes  map[string]any `json:"attributes"`
}
