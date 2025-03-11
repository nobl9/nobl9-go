package report

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type SystemHealthReviewConfig struct {
	TimeFrame  SystemHealthReviewTimeFrame `json:"timeFrame" validate:"required"`
	RowGroupBy RowGroupBy                  `json:"rowGroupBy" validate:"required" example:"project"`
	Columns    []ColumnSpec                `json:"columns" validate:"min=1,max=30"`
	Thresholds Thresholds                  `json:"thresholds" validate:"required"`
}

type Thresholds struct {
	RedLessThanOrEqual *float64 `json:"redLte" validate:"required" example:"0.8"`
	// Yellow is calculated as the difference between Red and Green
	// thresholds. If Red and Green are the same, Yellow is not used on the report.
	GreenGreaterThan *float64 `json:"greenGt" validate:"required" example:"0.95"`
	// ShowNoData customizes the report to either show or hide rows with no data.
	ShowNoData bool `json:"showNoData"`
}

type ColumnSpec struct {
	DisplayName string         `json:"displayName" validate:"required"`
	Labels      v1alpha.Labels `json:"labels" validate:"required"`
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
