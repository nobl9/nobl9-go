package report

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type SystemHealthReviewConfig struct {
	TimeFrame   SystemHealthReviewTimeFrame `json:"timeFrame"`
	RowGroupBy  RowGroupBy                  `json:"rowGroupBy"`
	Columns     []ColumnSpec                `json:"columns"`
	LabelRows   []LabelRowSpec              `json:"labelRows,omitempty"`
	Thresholds  Thresholds                  `json:"thresholds"`
	TableHeader string                      `json:"tableHeader,omitempty"`
}

type Thresholds struct {
	RedLessThanOrEqual *float64 `json:"redLte"`
	// Yellow is calculated as the difference between Red and Green
	// thresholds. If Red and Green are the same, Yellow is not used on the report.
	GreenGreaterThan *float64 `json:"greenGt"`
	// ShowNoData customizes the report to either show or hide rows with no data.
	ShowNoData bool `json:"showNoData"`
}

type ColumnSpec struct {
	DisplayName string         `json:"displayName"`
	Labels      v1alpha.Labels `json:"labels"`
}

type LabelRowSpec struct {
	DisplayName string         `json:"displayName,omitempty"`
	Labels      v1alpha.Labels `json:"labels"`
}

type SnapshotTimeFrame struct {
	Point    SnapshotPoint `json:"point"`
	DateTime *time.Time    `json:"dateTime,omitempty"`
	Rrule    string        `json:"rrule,omitempty"`
}

type SystemHealthReviewTimeFrame struct {
	Snapshot SnapshotTimeFrame `json:"snapshot"`
	TimeZone string            `json:"timeZone"`
}
