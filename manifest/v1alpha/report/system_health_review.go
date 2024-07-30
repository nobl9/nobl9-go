package report

import "time"

type SystemHealthReviewConfig struct {
	TimeFrame  SystemHealthReviewTimeFrame `json:"timeFrame" validate:"required"`
	RowGroupBy RowGroupBy                  `json:"rowGroupBy" validate:"required" example:"project"`
	Columns    []ColumnSpec                `json:"columns"`
}

type ColumnSpec struct {
	DisplayName string `json:"displayName" validate:"required"`
	Labels      Labels `json:"labels" validate:"required"`
}

type SnapshotTimeFrame struct {
	Point    SnapshotPoint `json:"point" validate:"required" example:"latest"`
	DateTime *time.Time    `json:"dateTime,omitempty"`
	Rrule    string        `json:"rrule,omitempty"`
}

type SystemHealthReviewTimeFrame struct {
	Snapshot SnapshotTimeFrame `json:"snapshot" validate:"required"`
	TimeZone string            `json:"timeZone" validate:"required,timeZone" example:"America/New_York"`
}
