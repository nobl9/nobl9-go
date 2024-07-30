package report

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Report '.*' has failed for the following fields:
.*
`))

func TestValidate_VersionAndKind(t *testing.T) {
	report := validReport()
	report.APIVersion = "v0.1"
	report.Kind = manifest.KindProject
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: validation.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: validation.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	report := validReport()
	report.Metadata = Metadata{
		Name: strings.Repeat("-my report", 20),
	}
	report.ManifestSource = "/home/me/report.yaml"
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 2,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringLength,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("fails with empty spec", func(t *testing.T) {
		report := validReport()
		report.Spec = Spec{}
		err := validate(report)
		testutils.AssertContainsErrors(t, report, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one report type configuration is required",
		})
	})
	t.Run("fails with more than one report type configuration defined in spec", func(t *testing.T) {
		report := validReport()
		report.Spec = Spec{
			SystemHealthReview: &SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					TimeZone: "Europe/Warsaw",
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointLatest,
					},
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{
						DisplayName: "Column 1",
						Labels: map[LabelKey][]LabelValue{
							"key1": {"value1"},
						},
					},
				},
			},
			SLOHistory: &SLOHistoryConfig{
				TimeFrame: SLOHistoryTimeFrame{
					TimeZone: "Europe/Warsaw",
					Rolling: &RollingTimeFrame{
						Repeat{
							Unit:  func(s string) *string { return &s }("day"),
							Count: func(i int) *int { return &i }(3),
						},
					},
				},
			},
		}
		err := validate(report)
		testutils.AssertContainsErrors(t, report, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one report type configuration is required",
		})
	})
}

func TestValidate_Spec_SLOHistory_TimeFrame(t *testing.T) {
	for name, timeFrame := range map[string]SLOHistoryTimeFrame{
		"passes with valid rolling time frame": {
			Rolling: &RollingTimeFrame{
				Repeat{
					Unit:  func(s string) *string { return &s }("day"),
					Count: func(i int) *int { return &i }(3),
				},
			},
			TimeZone: "Europe/Warsaw",
		},
		"passes with valid calendar repeating time frame": {
			Calendar: &CalendarTimeFrame{
				Repeat: Repeat{
					Unit:  func(s string) *string { return &s }("day"),
					Count: func(i int) *int { return &i }(3),
				},
			},
			TimeZone: "Europe/Warsaw",
		},
		"passes with valid calendar custom time frame": {
			Calendar: &CalendarTimeFrame{
				From: func(s string) *string { return &s }("2024-07-01"),
				To:   func(s string) *string { return &s }("2024-07-31"),
			},
			TimeZone: "Europe/Warsaw",
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validReport()
			report.Spec.SystemHealthReview = nil
			report.Spec.SLOHistory = &SLOHistoryConfig{
				TimeFrame: timeFrame,
			}
			err := validate(report)
			testutils.AssertNoError(t, report, err)
		})
	}
	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		TimeFrame           SLOHistoryTimeFrame
	}{
		"fails with empty rolling time frame": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.rolling.unit",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "spec.sloHistory.rolling.count",
					Code: validation.ErrorCodeRequired,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling:  &RollingTimeFrame{},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty count in rolling time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.rolling.count",
					Code: validation.ErrorCodeRequired,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit: func(s string) *string { return &s }("day"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty unit in rolling time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.rolling.unit",
					Code: validation.ErrorCodeRequired,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Count: func(i int) *int { return &i }(3),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty timeZone": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeZone",
					Code: validation.ErrorCodeRequired,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit:  func(s string) *string { return &s }("day"),
						Count: func(i int) *int { return &i }(3),
					},
				},
			},
		},
		"fails with invalid timeZone": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeZone",
					Message: "not a valid time zone: unknown time zone x",
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit:  func(s string) *string { return &s }("day"),
						Count: func(i int) *int { return &i }(3),
					},
				},
				TimeZone: "x",
			},
		},
		"fails with empty calendar time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.calendar",
					Message: "must contain either unit and count pair or from and to pair",
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{},
				TimeZone: "Europe/Warsaw",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validReport()
			report.Spec.SystemHealthReview = nil
			report.Spec.SLOHistory = &SLOHistoryConfig{
				TimeFrame: test.TimeFrame,
			}
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SystemHealthReview(t *testing.T) {
	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		Config              SystemHealthReviewConfig
	}{
		"fails with empty rowGroupBy value": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.rowGroupBy",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{Point: SnapshotPointLatest},
					TimeZone: "America/New_York",
				},
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with empty columns": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.columns",
					Code: validation.ErrorCodeSliceMinLength,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointLatest,
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns:    []ColumnSpec{},
			},
		},
		"fails with empty labels": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.columns[0].labels",
					Code: validation.ErrorCodeMapMinLength,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointLatest,
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{}},
				},
			},
		},
		"fails with empty displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.columns[0].displayName",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointLatest,
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validReport()
			report.Spec.SystemHealthReview = &test.Config
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SystemHealthReview_TimeFrame(t *testing.T) {
	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		Config              SystemHealthReviewConfig
	}{
		"fails with empty time frame": {
			ExpectedErrorsCount: 3,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.timeFrame",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "spec.systemHealthReview.timeZone",
					Code: validation.ErrorCodeRequired,
				},
				{
					Prop: "spec.systemHealthReview.snapshot",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with empty point in snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.snapshot",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with empty data in past point snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.snapshot.dateTime",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointPast,
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with wrong rrule format in past point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.snapshot.rrule",
					Message: "wrong format",
				},
				{
					Prop: "spec.systemHealthReview.snapshot.dateTime",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointPast,
						Rrule: "some test",
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with invalid rrule in past point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.snapshot.rrule",
					Message: "undefined frequency: TEST",
				},
				{
					Prop: "spec.systemHealthReview.snapshot.dateTime",
					Code: validation.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point: SnapshotPointPast,
						Rrule: "FREQ=TEST;INTERVAL=2",
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
		"fails with forbidden fields provided in latest point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.snapshot.rrule",
					Code: validation.ErrorCodeForbidden,
				},
				{
					Prop: "spec.systemHealthReview.snapshot.dateTime",
					Code: validation.ErrorCodeForbidden,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point:    SnapshotPointLatest,
						DateTime: ptr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
						Rrule:    "FREQ=DAY;INTERVAL=2",
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: map[LabelKey][]LabelValue{"key1": {"value1"}}},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validReport()
			report.Spec.SystemHealthReview = &test.Config
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func validReport() Report {
	return Report{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindReport,
		Metadata: Metadata{
			Name:        "my-report",
			DisplayName: "My Report",
		},
		Spec: Spec{
			Shared: true,
			Filters: &Filters{
				Projects: []Project{
					{
						Name: "project",
					},
				},
				Services: []Service{
					{
						Name:    "service",
						Project: "project",
					},
				},
				SLOs: []SLO{
					{
						Name:    "slo1",
						Project: "project",
						Service: "service",
					},
				},
				Labels: map[string][]string{
					"key1": {
						"value1",
						"value2",
					},
					"key2": {
						"value1",
						"value2",
					},
				},
			},
			SystemHealthReview: &SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point:    SnapshotPointPast,
						DateTime: ptr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
						Rrule:    "FREQ=WEEKLY",
					},
					TimeZone: "America/New_York",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{
						DisplayName: "Column 1",
						Labels: map[string][]string{
							"key1": {
								"value1",
							},
							"key2": {
								"value1",
								"value2",
							},
						},
					},
					{
						DisplayName: "Column 2",
						Labels: map[string][]string{
							"key3": {
								"value1",
							},
						},
					},
				},
			},
		},
	}
}

func ptr[T any](v T) *T { return &v }
