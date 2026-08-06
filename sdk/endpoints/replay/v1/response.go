package v1

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

// ReplayWithStatus identifies a replay and includes its current status.
type ReplayWithStatus struct {
	Project string       `json:"project"`
	SLO     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

// ReplayStatus describes the current state of a replay.
type ReplayStatus struct {
	Source       ReplaySource             `json:"source"`
	Status       ReplayListStatus         `json:"status"`
	Cancellation ReplayCancellationStatus `json:"cancellation"`
	CanceledBy   string                   `json:"canceledBy,omitempty"`
	TriggeredBy  string                   `json:"triggeredBy"`
	Unit         DurationUnit             `json:"unit"`
	StartTime    string                   `json:"startTime"`
	EndTime      string                   `json:"endTime,omitempty"`
	Value        int                      `json:"value"`
}

// ToProcessStatus converts ReplayStatus to the SLO manifest process status.
func (s ReplayStatus) ToProcessStatus() v1alphaSLO.ProcessStatus {
	return v1alphaSLO.ProcessStatus{
		Status:       string(s.Status),
		Cancellation: string(s.Cancellation),
		CanceledBy:   s.CanceledBy,
		TriggeredBy:  s.TriggeredBy,
		Unit:         string(s.Unit),
		Value:        s.Value,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
	}
}

// ReplayAvailability reports whether a replay can be started.
type ReplayAvailability struct {
	Reason    ReplayAvailabilityReason `json:"reason,omitempty"`
	Available bool                     `json:"available"`
}

// ReplayListItem summarizes an active replay.
// CreatedAt, when present, is an RFC3339 timestamp.
type ReplayListItem struct {
	SLO            string                   `json:"slo,omitempty"`
	Project        string                   `json:"project"`
	CreatedAt      string                   `json:"createdAt,omitempty"`
	ElapsedTime    string                   `json:"elapsedTime,omitempty"`
	RetrievedScope string                   `json:"retrievedScope,omitempty"`
	RetrievedFrom  string                   `json:"retrievedFrom,omitempty"`
	Status         ReplayListStatus         `json:"status"`
	Cancellation   ReplayCancellationStatus `json:"cancellation,omitempty"`
}
