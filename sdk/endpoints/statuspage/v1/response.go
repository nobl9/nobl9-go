package v1

import "time"

// ComponentStatus is the computed status of a status page component.
// Disruption severities share the same values, except [ComponentStatusOperational].
type ComponentStatus = string

const (
	ComponentStatusOperational         ComponentStatus = "operational"
	ComponentStatusDegradedPerformance ComponentStatus = "degradedPerformance"
	ComponentStatusMajorOutage         ComponentStatus = "majorOutage"
)

type GetStatusResponse struct {
	Components []StatusComponent `json:"components"`
}

// StatusComponent is a node in the status page component tree.
// Children makes the type self-referential by design — consumers which cannot
// handle recursive schemas (e.g. JSON schema generators) must flatten it themselves.
type StatusComponent struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	Status              ComponentStatus    `json:"status"`
	NoSignal            bool               `json:"noSignal"`
	SLOs                []SLOReference     `json:"slos,omitempty"`
	ImpactingDisruption *DisruptionDetail  `json:"impactingDisruption,omitempty"`
	DisruptionHistory   *DisruptionHistory `json:"disruptionHistory,omitempty"`
	Children            []StatusComponent  `json:"children,omitempty"`
}

// DisruptionDetail describes an ongoing disruption impacting a component.
type DisruptionDetail struct {
	StartedAt  time.Time `json:"startedAt"`
	Severity   string    `json:"severity"`
	IssueCount int       `json:"issueCount"`
	// Duration is the number of seconds since the disruption started.
	Duration int64 `json:"duration"`
}

type SLOReference struct {
	Project     string `json:"project"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

// DisruptionHistory summarizes a component's recent disruptions per day.
type DisruptionHistory struct {
	Days    []DayDisruptionCount     `json:"days"`
	Summary DisruptionHistorySummary `json:"summary"`
}

type DayDisruptionCount struct {
	Date                     string `json:"date"`
	Disruptions              int    `json:"disruptions"`
	DegradedPerformanceCount int    `json:"degradedPerformanceCount"`
	MajorOutageCount         int    `json:"majorOutageCount"`
}

type DisruptionHistorySummary struct {
	Total               int     `json:"total"`
	AvgPerDay           float64 `json:"avgPerDay"`
	DaysWithDisruptions int     `json:"daysWithDisruptions"`
}

type ListDisruptionsResponse struct {
	Disruptions []Disruption `json:"disruptions"`
	Total       int64        `json:"total"`
	Limit       int32        `json:"limit"`
	Offset      int32        `json:"offset"`
}

type Disruption struct {
	ID                 string                   `json:"id"`
	Title              *string                  `json:"title,omitempty"`
	Severity           string                   `json:"severity"`
	Source             string                   `json:"source"`
	IsCleared          bool                     `json:"isCleared"`
	OriginComponent    ComponentReference       `json:"originComponent"`
	AffectedComponents []ComponentReference     `json:"affectedComponents"`
	StartTime          time.Time                `json:"startTime"`
	EndTime            *time.Time               `json:"endTime,omitempty"`
	History            []DisruptionHistoryEvent `json:"history"`
	CreatedBy          string                   `json:"createdBy"`
	CreatedAt          time.Time                `json:"createdAt"`
	Metadata           map[string]any           `json:"metadata,omitempty"`
}

type ComponentReference struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted,omitempty"`
}

// DisruptionHistoryEvent is a single status change in a disruption's lifecycle.
type DisruptionHistoryEvent struct {
	ID             string    `json:"id"`
	PreviousStatus string    `json:"previousStatus"`
	NewStatus      string    `json:"newStatus"`
	ChangedBy      string    `json:"changedBy"`
	ChangedAt      time.Time `json:"changedAt"`
	Comment        string    `json:"comment,omitempty"`
	Source         string    `json:"source"`
}
