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
	Status           AlertAnalysisStatus     `json:"status"`
	DetectionStatus  AnalysisReadinessStatus `json:"detectionStatus"`
	TimeseriesStatus AnalysisReadinessStatus `json:"timeseriesStatus"`
	Timeseries       []AlertingTimeseries    `json:"timeseries,omitempty"`
}

type AlertAnalysisStatus string

const (
	AlertAnalysisStatusExportingTimeseries      AlertAnalysisStatus = "exporting_timeseries"
	AlertAnalysisStatusCalculatingAlertsmetrics AlertAnalysisStatus = "calculating_alertsmetrics"
	AlertAnalysisStatusCalculatingAlerts        AlertAnalysisStatus = "calculating_alerts"
	AlertAnalysisStatusStoreAlerts              AlertAnalysisStatus = "store_alerts"
	AlertAnalysisStatusCreatingVersion          AlertAnalysisStatus = "creating_version"
	AlertAnalysisStatusDownsampling             AlertAnalysisStatus = "downsampling"
	AlertAnalysisStatusUpdatingTimeseries       AlertAnalysisStatus = "updating_timeseries"
	AlertAnalysisStatusEnablingVersion          AlertAnalysisStatus = "enabling_version"
	AlertAnalysisStatusDone                     AlertAnalysisStatus = "done"
	AlertAnalysisStatusCanceled                 AlertAnalysisStatus = "canceled"
	AlertAnalysisStatusError                    AlertAnalysisStatus = "error"
)

type AnalysisReadinessStatus string

const (
	AnalysisReadinessStatusPending  AnalysisReadinessStatus = "pending"
	AnalysisReadinessStatusRunning  AnalysisReadinessStatus = "running"
	AnalysisReadinessStatusReady    AnalysisReadinessStatus = "ready"
	AnalysisReadinessStatusCanceled AnalysisReadinessStatus = "canceled"
	AnalysisReadinessStatusError    AnalysisReadinessStatus = "error"
)

type AlertingTimeseries struct {
	Measurement string         `json:"measurement"`
	Timestamps  []int64        `json:"timestamps"`
	Values      []float64      `json:"values"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}
