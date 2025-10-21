package report

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Report '.*' has failed for the following fields:
.*
`))

func TestValidate_VersionAndKind(t *testing.T) {
	report := validSystemHealthReport()
	report.APIVersion = "v0.1"
	report.Kind = manifest.KindProject
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	report := validSystemHealthReport()
	report.Metadata = Metadata{
		Name: strings.Repeat("-my report", 20),
	}
	report.ManifestSource = "/home/me/report.yaml"
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 1,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("fails with empty spec", func(t *testing.T) {
		report := validSystemHealthReport()
		report.Spec = Spec{}
		err := validate(report)
		testutils.AssertContainsErrors(t, report, err, 2,
			testutils.ExpectedError{
				Prop:    "spec",
				Message: "exactly one report type configuration is required",
			},
			testutils.ExpectedError{
				Prop: "spec.filters",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("fails with more than one report type configuration defined in spec", func(t *testing.T) {
		report := validSystemHealthReport()
		report.Spec = Spec{
			Filters: &Filters{Projects: []string{"project"}},
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
						Labels: v1alpha.Labels{
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
							Unit:  func(s string) *string { return &s }("Week"),
							Count: func(i int) *int { return &i }(1),
						},
					},
				},
			},
		}
		err := validate(report)
		testutils.AssertContainsErrors(t, report, err, 2, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one report type configuration is required",
		})
	})
}

func TestValidate_Spec_Filters(t *testing.T) {
	for name, filters := range map[string]Filters{
		"passes with valid projects": {
			Projects: []string{"project1", "project2"},
		},
		"passes with valid services": {
			Services: Services{
				Service{
					Name:    "service1",
					Project: "project1",
				},
				Service{
					Name:    "service2",
					Project: "project2",
				},
			},
		},
		"passes with valid SLOs": {
			SLOs: SLOs{
				SLO{
					Name:    "slo1",
					Project: "project1",
				},
				SLO{
					Name:    "slo2",
					Project: "project2",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			report.Spec.Filters = &filters
			err := validate(report)
			testutils.AssertNoError(t, report, err)
		})
	}
	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		Filters             *Filters
	}{
		"fails with empty filters": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters",
					Code: rules.ErrorCodeRequired,
				},
			},
			Filters: nil,
		},
		"fails with neither projects, services nor slos selected": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.filters",
					Message: "at least one of the following fields is required: projects, services, slos",
				},
			},
			Filters: &Filters{
				Labels: v1alpha.Labels{
					"key1": {"value1"},
				},
			},
		},
		"fails with invalid project names": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.projects[0]",
					Code: validationV1Alpha.ErrorCodeStringName,
				},
			},
			Filters: &Filters{
				Projects: []string{"test project", "project"},
			},
		},
		"fails with service defined without name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.services[0].name",
					Code: rules.ErrorCodeRequired,
				},
			},
			Filters: &Filters{
				Services: Services{
					Service{
						Project: "project",
					},
				},
			},
		},
		"fails with service defined with invalid name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.services[0].name",
					Code: validationV1Alpha.ErrorCodeStringName,
				},
			},
			Filters: &Filters{
				Services: Services{
					Service{
						Name:    "test name",
						Project: "project",
					},
				},
			},
		},
		"fails with service defined without project": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.services[0].project",
					Code: rules.ErrorCodeRequired,
				},
			},
			Filters: &Filters{
				Services: Services{
					Service{
						Name: "service",
					},
				},
			},
		},
		"fails with service defined with invalid project": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.services[0].project",
					Code: validationV1Alpha.ErrorCodeStringName,
				},
			},
			Filters: &Filters{
				Services: Services{
					Service{
						Name:    "name",
						Project: "test project",
					},
				},
			},
		},
		"fails with slo defined without name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].name",
					Code: rules.ErrorCodeRequired,
				},
			},
			Filters: &Filters{
				SLOs: SLOs{
					SLO{
						Project: "project",
					},
				},
			},
		},
		"fails with slo defined with invalid name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].name",
					Code: validationV1Alpha.ErrorCodeStringName,
				},
			},
			Filters: &Filters{
				SLOs: SLOs{
					SLO{
						Name:    "test name",
						Project: "project",
					},
				},
			},
		},
		"fails with slo defined without project": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].project",
					Code: rules.ErrorCodeRequired,
				},
			},
			Filters: &Filters{
				SLOs: SLOs{
					SLO{
						Name: "service",
					},
				},
			},
		},
		"fails with slo defined with invalid project": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].project",
					Code: validationV1Alpha.ErrorCodeStringName,
				},
			},
			Filters: &Filters{
				SLOs: SLOs{
					SLO{
						Name:    "name",
						Project: "test project",
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			report.Spec.Filters = test.Filters
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SLOHistory_TimeFrame(t *testing.T) {
	validUnitAndCountRollingPairs := "valid 'unit' and 'count' pairs are: " +
		"1 Week, 2 Week, 4 Week, 1 Month, 1 Quarter, 1 Year"
	validUnitAndCountCalendarPairs := "valid 'unit' and 'count' pairs are: 1 Week, 1 Month, 1 Quarter, 1 Year"
	validCalendarPairs := "must contain either 'unit' and 'count' pair or 'from' and 'to' pair"
	futureDate := time.Now().Add(time.Hour * 24)

	for name, timeFrame := range map[string]SLOHistoryTimeFrame{
		"passes with valid rolling time frame": {
			Rolling: &RollingTimeFrame{
				Repeat{
					Unit:  func(s string) *string { return &s }("Week"),
					Count: func(i int) *int { return &i }(1),
				},
			},
			TimeZone: "Europe/Warsaw",
		},
		"passes with valid calendar repeating time frame": {
			Calendar: &CalendarTimeFrame{
				Repeat: Repeat{
					Unit:  func(s string) *string { return &s }("Quarter"),
					Count: func(i int) *int { return &i }(1),
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
			report := validSystemHealthReport()
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
			ExpectedErrorsCount: 3,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.rolling.unit",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop: "spec.sloHistory.timeFrame.rolling.count",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop:    "spec.sloHistory.timeFrame.rolling",
					Message: validUnitAndCountRollingPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling:  &RollingTimeFrame{},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty count in rolling time frame": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.rolling.count",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop:    "spec.sloHistory.timeFrame.rolling",
					Message: validUnitAndCountRollingPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit: func(s string) *string { return &s }("Week"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty unit in rolling time frame": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.rolling.unit",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop:    "spec.sloHistory.timeFrame.rolling",
					Message: validUnitAndCountRollingPairs,
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
		"fails with wrong unit in rolling time frame": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.rolling.unit",
					Code: rules.ErrorCodeOneOf,
				},
				{
					Prop:    "spec.sloHistory.timeFrame.rolling",
					Message: validUnitAndCountRollingPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Count: func(i int) *int { return &i }(3),
						Unit:  func(s string) *string { return &s }("Day"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with wrong unit and count pair in rolling time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.rolling",
					Message: validUnitAndCountRollingPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Count: func(i int) *int { return &i }(3),
						Unit:  func(s string) *string { return &s }("Week"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with empty timeZone": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.timeZone",
					Code: rules.ErrorCodeRequired,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit:  func(s string) *string { return &s }("Week"),
						Count: func(i int) *int { return &i }(1),
					},
				},
			},
		},
		"fails with invalid timeZone": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.timeZone",
					Message: "not a valid time zone: unknown time zone x",
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Rolling: &RollingTimeFrame{
					Repeat{
						Unit:  func(s string) *string { return &s }("Week"),
						Count: func(i int) *int { return &i }(1),
					},
				},
				TimeZone: "x",
			},
		},
		"fails with empty calendar time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.calendar",
					Message: validCalendarPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with half empty calendar time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.calendar",
					Message: validCalendarPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{
					Repeat: Repeat{
						Count: func(i int) *int { return &i }(1),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with wrong unit in calendar time frame": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.sloHistory.timeFrame.calendar.unit",
					Code: rules.ErrorCodeOneOf,
				},
				{
					Prop:    "spec.sloHistory.timeFrame.calendar",
					Message: validUnitAndCountCalendarPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{
					Repeat: Repeat{
						Count: func(i int) *int { return &i }(3),
						Unit:  func(s string) *string { return &s }("Day"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with wrong unit and count pair in calendar time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.calendar",
					Message: validUnitAndCountCalendarPairs,
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{
					Repeat: Repeat{
						Count: func(i int) *int { return &i }(3),
						Unit:  func(s string) *string { return &s }("Week"),
					},
				},
				TimeZone: "Europe/Warsaw",
			},
		},
		"fails with dates in the past in calendar time frame": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.sloHistory.timeFrame.calendar",
					Message: "dates must be in the past",
				},
			},
			TimeFrame: SLOHistoryTimeFrame{
				Calendar: &CalendarTimeFrame{
					From: func(s string) *string { return &s }("2024-01-01"),
					To:   func(s string) *string { return &s }(futureDate.Format(time.DateOnly)),
				},
				TimeZone: "Europe/Warsaw",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
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
	t.Run("valid rowGroupBy values", func(t *testing.T) {
		for _, rowGroupBy := range RowGroupByValues() {
			report := validSystemHealthReport()
			config := validSystemHealthReviewConfig(rowGroupBy)
			report.Spec.SystemHealthReview = &config
			err := validate(report)
			testutils.AssertNoError(t, report, err)
		}
	})

	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		ConfigFunc          func(conf SystemHealthReviewConfig) SystemHealthReviewConfig
	}{
		"fails with empty rowGroupBy value": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.rowGroupBy",
				Code: rules.ErrorCodeRequired,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.RowGroupBy = RowGroupBy(0)
				return conf
			},
		},
		"fails with empty thresholds": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.thresholds",
				Code: rules.ErrorCodeRequired,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Thresholds = Thresholds{}
				return conf
			},
		},
		"fails with invalid thresholds": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.thresholds.greenGt",
				Code: rules.ErrorCodeLessThan,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Thresholds = Thresholds{
					RedLessThanOrEqual: ptr(-0.1),
					GreenGreaterThan:   ptr(1.1),
				}
				return conf
			},
		},
		"fails when red is greater than green": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop:    "spec.systemHealthReview.thresholds.redLte",
				Message: "must be less than or equal to 'greenGt' (0.1)",
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Thresholds = Thresholds{
					RedLessThanOrEqual: ptr(0.2),
					GreenGreaterThan:   ptr(0.1),
				}
				return conf
			},
		},
		"fails when rowGroupBy is 'project'": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeForbidden,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.RowGroupBy = RowGroupByProject
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{"key1": {"value1"}}}}
				return conf
			},
		},
		"fails when rowGroupBy is 'service'": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeForbidden,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.RowGroupBy = RowGroupByService
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{"key1": {"value1"}}}}
				return conf
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			conf := validSystemHealthReviewConfig(RowGroupByService)
			conf = test.ConfigFunc(conf)
			report.Spec.SystemHealthReview = &conf
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SystemHealthReview_Columns(t *testing.T) {
	for name, test := range map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		ConfigFunc          func(conf SystemHealthReviewConfig) SystemHealthReviewConfig
	}{
		"fails with empty columns": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.columns",
				Code: rules.ErrorCodeSliceLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Columns = []ColumnSpec{}
				return conf
			},
		},
		"fails with too many columns": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.columns",
				Code: rules.ErrorCodeSliceLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				limit := 31
				result := make([]ColumnSpec, limit)
				for i := range limit {
					result[i] = ColumnSpec{DisplayName: strconv.Itoa(i), Labels: v1alpha.Labels{"key1": {"value1"}}}
				}
				conf.Columns = result
				return conf
			},
		},
		"fails with empty labels": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.columns[0].labels",
				Code: rules.ErrorCodeMapMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Columns[0].Labels = v1alpha.Labels{}
				return conf
			},
		},
		"fails with invalid label key": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop:       "spec.systemHealthReview.columns[0].labels.['k ey']",
				IsKeyError: true,
				Code:       rules.ErrorCodeStringMatchRegexp,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Columns[0].Labels = v1alpha.Labels{"k ey": nil}
				return conf
			},
		},
		"fails with empty displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.columns[0].displayName",
				Code: rules.ErrorCodeRequired,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Columns[0].DisplayName = ""
				return conf
			},
		},
		"fails with too long displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.columns[0].displayName",
				Code: rules.ErrorCodeStringMaxLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.Columns[0].DisplayName = strings.Repeat("l", 254)
				return conf
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			conf := validSystemHealthReviewConfig(RowGroupByProject)
			conf = test.ConfigFunc(conf)
			report.Spec.SystemHealthReview = &conf
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SystemHealthReview_RowGroupByLabel(t *testing.T) {
	t.Run("passes with valid labelRows when rowGroupBy is label", func(t *testing.T) {
		report := validSystemHealthReport()
		conf := validSystemHealthReviewConfig(RowGroupByLabel)
		conf.RowGroupBy = RowGroupByLabel
		conf.LabelRows = []LabelRowSpec{
			{Labels: v1alpha.Labels{"env": nil}},
		}
		report.Spec.SystemHealthReview = &conf
		err := validate(report)
		testutils.AssertNoError(t, report, err)
	})

	tests := map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		ConfigFunc          func(conf SystemHealthReviewConfig) SystemHealthReviewConfig
	}{
		"fails with nil labelRows": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeSliceLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = nil
				return conf
			},
		},
		"fails with empty labelRows": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeSliceLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{}
				return conf
			},
		},
		"fails with too many labelRows": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeSliceLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				limit := 2
				result := make([]LabelRowSpec, limit)
				for i := range limit {
					result[i] = LabelRowSpec{Labels: v1alpha.Labels{"key" + strconv.Itoa(i): nil}}
				}
				conf.LabelRows = result
				return conf
			},
		},
		"fails with empty labels": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].labels",
				Code: rules.ErrorCodeMapLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{}}}
				return conf
			},
		},
		"fails with too many labels": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].labels",
				Code: rules.ErrorCodeMapLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{"key1": nil, "key2": nil}}}
				return conf
			},
		},
		"fails with invalid label key": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop:       "spec.systemHealthReview.labelRows[0].labels.['k ey']",
				IsKeyError: true,
				Code:       rules.ErrorCodeStringMatchRegexp,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{"k ey": nil}}}
				return conf
			},
		},
		"fails with label values": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop:    "spec.systemHealthReview.labelRows[0].labels.key1",
				Message: "label values must be empty",
				Code:    rules.ErrorCodeSliceMaxLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{{Labels: v1alpha.Labels{"key1": {"value1"}}}}
				return conf
			},
		},
		"fails with displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].displayName",
				Code: rules.ErrorCodeForbidden,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].DisplayName = "foo"
				return conf
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			conf := validSystemHealthReviewConfig(RowGroupByLabel)
			conf = test.ConfigFunc(conf)
			report.Spec.SystemHealthReview = &conf
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_SystemHealthReview_RowGroupByCustom(t *testing.T) {
	t.Run("passes with valid labelRows when rowGroupBy is custom", func(t *testing.T) {
		report := validSystemHealthReport()
		conf := validSystemHealthReviewConfig(RowGroupByCustom)
		conf.RowGroupBy = RowGroupByCustom
		conf.LabelRows = []LabelRowSpec{{
			DisplayName: "Environment",
			Labels:      v1alpha.Labels{"env": []string{"prod"}},
		}}
		report.Spec.SystemHealthReview = &conf
		err := validate(report)
		testutils.AssertNoError(t, report, err)
	})

	tests := map[string]struct {
		ExpectedErrorsCount int
		ExpectedErrors      []testutils.ExpectedError
		ConfigFunc          func(conf SystemHealthReviewConfig) SystemHealthReviewConfig
	}{
		"fails with nil labelRows": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeSliceMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = nil
				return conf
			},
		},
		"fails with empty labelRows": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows",
				Code: rules.ErrorCodeSliceMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows = []LabelRowSpec{}
				return conf
			},
		},
		"fails with empty labels": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].labels",
				Code: rules.ErrorCodeMapMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].Labels = v1alpha.Labels{}
				return conf
			},
		},
		"fails with invalid label key": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop:       "spec.systemHealthReview.labelRows[0].labels.['k ey']",
				IsKeyError: true,
				Code:       rules.ErrorCodeStringMatchRegexp,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].Labels = v1alpha.Labels{"k ey": nil}
				return conf
			},
		},
		"fails with empty label values": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].labels.key1",
				Code: rules.ErrorCodeSliceMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].Labels = v1alpha.Labels{"key1": []string{}}
				return conf
			},
		},
		"fails with nil label values": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].labels.key1",
				Code: rules.ErrorCodeSliceMinLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].Labels = v1alpha.Labels{"key1": nil}
				return conf
			},
		},
		"fails with empty displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].displayName",
				Code: rules.ErrorCodeRequired,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].DisplayName = ""
				return conf
			},
		},
		"fails with too long displayName": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{{
				Prop: "spec.systemHealthReview.labelRows[0].displayName",
				Code: rules.ErrorCodeStringMaxLength,
			}},
			ConfigFunc: func(conf SystemHealthReviewConfig) SystemHealthReviewConfig {
				conf.LabelRows[0].DisplayName = strings.Repeat("l", validationV1Alpha.NameMaximumLength+1)
				return conf
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			conf := validSystemHealthReviewConfig(RowGroupByCustom)
			conf = test.ConfigFunc(conf)
			report.Spec.SystemHealthReview = &conf
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
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.timeFrame",
					Code: rules.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with empty point in snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot",
					Code: rules.ErrorCodeRequired,
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with empty data in past point snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot.dateTime",
					Code: rules.ErrorCodeRequired,
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
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with wrong rrule format in past point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.timeFrame.snapshot.rrule",
					Message: "wrong format",
				},
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot.dateTime",
					Code: rules.ErrorCodeRequired,
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
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with invalid rrule in past point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.timeFrame.snapshot.rrule",
					Message: "undefined frequency: TEST",
				},
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot.dateTime",
					Code: rules.ErrorCodeRequired,
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
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with rrule frequency less than DAILY in past point snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.timeFrame.snapshot.rrule",
					Message: "rrule must have at least daily frequency",
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point:    SnapshotPointPast,
						DateTime: ptr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
						Rrule:    "FREQ=SECONDLY;INTERVAL=2",
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with forbidden fields provided in latest point snapshot": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot.rrule",
					Code: rules.ErrorCodeForbidden,
				},
				{
					Prop: "spec.systemHealthReview.timeFrame.snapshot.dateTime",
					Code: rules.ErrorCodeForbidden,
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
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
		"fails with dateTime in the future in past point snapshot": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.systemHealthReview.timeFrame.snapshot.dateTime",
					Message: "dateTime must be in the past",
				},
			},
			Config: SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point:    SnapshotPointPast,
						DateTime: ptr(time.Now().Add(time.Hour)),
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
				},
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.0),
					GreenGreaterThan:   ptr(0.2),
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			report := validSystemHealthReport()
			report.Spec.SystemHealthReview = &test.Config
			err := validate(report)
			testutils.AssertContainsErrors(t, report, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestAtLeastDailyFreq(t *testing.T) {
	atLeastDailyFreqErr := "rrule must have at least daily frequency"
	tests := []struct {
		name          string
		rule          string
		expectedError string
	}{
		{
			name:          "nil rule returns no error",
			rule:          "",
			expectedError: "",
		},
		{
			name:          "hourly frequency returns error",
			rule:          "FREQ=HOURLY;INTERVAL=1",
			expectedError: atLeastDailyFreqErr,
		},
		{
			name:          "minutely frequency returns error",
			rule:          "FREQ=MINUTELY;INTERVAL=1",
			expectedError: atLeastDailyFreqErr,
		},
		{
			name:          "secondly frequency returns error",
			rule:          "FREQ=SECONDLY;INTERVAL=1",
			expectedError: atLeastDailyFreqErr,
		},
		{
			name:          "daily frequency returns no error",
			rule:          "FREQ=DAILY;INTERVAL=59",
			expectedError: "",
		},
		{
			name:          "weekly frequency returns no error",
			rule:          "FREQ=WEEKLY;INTERVAL=59",
			expectedError: "",
		},
		{
			name:          "monthly frequency returns no error",
			rule:          "FREQ=MONTHLY;INTERVAL=59",
			expectedError: "",
		},
		{
			name:          "yearly frequency returns no error",
			rule:          "FREQ=YEARLY;INTERVAL=59",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule *rrule.RRule
			if tt.rule != "" {
				rule, _ = rrule.StrToRRule(tt.rule)
			}
			err := atLeastDailyFreq.Validate(rule)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func validSystemHealthReport() Report {
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
				Projects: []string{"project"},
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
				Thresholds: Thresholds{
					RedLessThanOrEqual: ptr(0.8),
					GreenGreaterThan:   ptr(0.95),
					ShowNoData:         true,
				},
			},
		},
	}
}

func validSystemHealthReviewConfig(rowGroupBy RowGroupBy) SystemHealthReviewConfig {
	config := SystemHealthReviewConfig{
		TimeFrame: SystemHealthReviewTimeFrame{
			Snapshot: SnapshotTimeFrame{
				Point: SnapshotPointLatest,
			},
			TimeZone: "Europe/Warsaw",
		},
		RowGroupBy: rowGroupBy,
		Columns: []ColumnSpec{
			{DisplayName: "Column 1", Labels: v1alpha.Labels{"key1": {"value1"}}},
		},
		Thresholds: Thresholds{
			RedLessThanOrEqual: ptr(0.0),
			GreenGreaterThan:   ptr(0.2),
		},
	}
	// nolint: exhaustive
	switch rowGroupBy {
	case RowGroupByLabel:
		config.LabelRows = []LabelRowSpec{{
			Labels: v1alpha.Labels{"environment": nil},
		}}
	case RowGroupByCustom:
		config.LabelRows = []LabelRowSpec{
			{
				DisplayName: "Production databases",
				Labels:      v1alpha.Labels{"environment": []string{"prod"}, "type": []string{"database"}},
			},
			{
				DisplayName: "R&D",
				Labels:      v1alpha.Labels{"team": []string{"green", "gray"}},
			},
			{
				DisplayName: "Critical",
				Labels:      v1alpha.Labels{"level": []string{"critical"}},
			},
		}
	}
	return config
}
