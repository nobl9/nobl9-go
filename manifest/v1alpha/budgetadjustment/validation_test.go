package budgetadjustment

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/rrule-go"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for BudgetAdjustment '.*' has failed for the following fields:
.*
Manifest source: /home/me/budget-adjustment.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	budgetAdjustment := validBudgetAdjustment()
	budgetAdjustment.APIVersion = "v0.1"
	budgetAdjustment.Kind = manifest.KindProject
	budgetAdjustment.ManifestSource = "/home/me/budget-adjustment.yaml"
	err := validate(budgetAdjustment)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, budgetAdjustment, err, 2,
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
	budgetAdjustment := validBudgetAdjustment()
	budgetAdjustment.Metadata = Metadata{
		Name:        strings.Repeat("MY BUDGET ADJUSTMENT ", 20),
		DisplayName: strings.Repeat("my-budget-adjustment-", 20),
	}
	budgetAdjustment.ManifestSource = "/home/me/budget-adjustment.yaml"
	err := validate(budgetAdjustment)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, budgetAdjustment, err, 2,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedErrors []testutils.ExpectedError
	}{
		{
			name: "description too long",
			spec: Spec{
				Description:     strings.Repeat("A", 2000),
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters:         Filters{SLOs: []SLORef{{Name: "my-slo", Project: "default"}}},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.description",
					Code: validationV1Alpha.ErrorCodeStringDescription,
				},
			},
		},
		{
			name: "first event start required",
			spec: Spec{
				Duration: "1m",
				Filters:  Filters{SLOs: []SLORef{{Name: "my-slo", Project: "default"}}},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.firstEventStart",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		{
			name: "duration required",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Filters:         Filters{SLOs: []SLORef{{Name: "my-slo", Project: "default"}}},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.duration",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		{
			name: "no slo filters",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters:         Filters{},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.filters.slos",
					Message: "length must be greater than or equal to 1",
				},
			},
		},
		{
			name: "too short duration",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1s",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "test",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.duration",
					Code: rules.ErrorCodeDurationPrecision,
				},
			},
		},
		{
			name: "duration contains seconds",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m1s",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "test",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.duration",
					Code: rules.ErrorCodeDurationPrecision,
				},
			},
		},
		{
			name: "slo is defined without name",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters: Filters{
					SLOs: []SLORef{{
						Project: "test",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].name",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		{
			name: "slo is defined with invalid slo name",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "Test name",
						Project: "test",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].name",
					Code: rules.ErrorCodeStringDNSLabel,
				},
			},
		},
		{
			name: "slo is defined without project",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters: Filters{
					SLOs: []SLORef{{
						Name: "test",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].project",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		{
			name: "slo is defined with invalid project name",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "name",
						Project: "Project name",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos[0].project",
					Code: rules.ErrorCodeStringDNSLabel,
				},
			},
		},
		{
			name: "wrong rrule format",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "some test",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.rrule",
					Message: "wrong format",
				},
			},
		},
		{
			name: "invalid rrule",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=TEST;INTERVAL=2",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.rrule",
					Message: "undefined frequency: TEST",
				},
			},
		},
		{
			name: "invalid freq in rrule",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=MINUTELY;INTERVAL=2;COUNT=10",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.rrule",
					Message: "interval must be at least 60 minutes for minutely frequency",
				},
			},
		},
		{
			name: "rrule with dtstart trows transform error",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "DTSTART:20240909T065900Z\\nRRULE:FREQ=MINUTELY;BYHOUR=6,8,9,10,11,12,13,14,15,16;COUNT=10",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.rrule",
					Code: "transform",
				},
			},
		},
		{
			name: "proper freq in rrule",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=HOURLY;INTERVAL=1",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedErrors: []testutils.ExpectedError{},
		},
		{
			name: "duplicate slo",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=WEEKLY;INTERVAL=2",
				Filters: Filters{
					SLOs: []SLORef{
						{Name: "test", Project: "project"},
						{Name: "test", Project: "project"},
					},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos",
					Code: rules.ErrorCodeSliceUnique,
				},
			},
		},
		{
			name: "duplicate slo with multiple others",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=WEEKLY;INTERVAL=2",
				Filters: Filters{
					SLOs: []SLORef{
						{Name: "test1", Project: "project"},
						{Name: "test1", Project: "project"},
						{Name: "test2", Project: "project"},
						{Name: "test3", Project: "project"},
						{Name: "test4", Project: "project"},
						{Name: "test5", Project: "project"},
						{Name: "test6", Project: "project"},
					},
				},
			},
			expectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.filters.slos",
					Code: rules.ErrorCodeSliceUnique,
				},
			},
		},
		{
			name: "proper spec",
			spec: Spec{
				FirstEventStart: time.Now().Truncate(time.Second),
				Duration:        "1m",
				Rrule:           "FREQ=WEEKLY;INTERVAL=2",
				Filters: Filters{
					SLOs: []SLORef{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			budgetAdjustment := validBudgetAdjustment()
			budgetAdjustment.Spec = test.spec

			err := validate(budgetAdjustment)

			if len(test.expectedErrors) == 0 {
				testutils.AssertNoError(t, test.spec, err)
			} else {
				testutils.AssertContainsErrors(t, test.spec, err, len(test.expectedErrors), test.expectedErrors...)
			}
		})
	}
}

func validBudgetAdjustment() BudgetAdjustment {
	return BudgetAdjustment{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindBudgetAdjustment,
		Metadata: Metadata{
			Name:        "my-budget-adjustment",
			DisplayName: "My Budget Adjustment",
		},
		Spec: Spec{
			FirstEventStart: time.Now().Truncate(time.Second),
			Duration:        "1m",
			Filters: Filters{
				SLOs: []SLORef{
					{
						Name:    "my-slo",
						Project: "default",
					},
				},
			},
		},
	}
}

func TestAtLeastHourlyFreq(t *testing.T) {
	tests := []struct {
		name          string
		rule          string
		expectedError string
	}{
		{
			name:          "nil rule returns nil error",
			rule:          "",
			expectedError: "",
		},
		{
			name:          "frequency less than hourly returns error",
			rule:          "FREQ=MINUTELY;INTERVAL=1",
			expectedError: "interval must be at least 60 minutes for minutely frequency",
		},
		{
			name:          "frequency less than hourly returns error",
			rule:          "FREQ=MINUTELY;INTERVAL=59;COUNT=10",
			expectedError: "interval must be at least 60 minutes for minutely frequency",
		},
		{
			name:          "frequency less than hourly for single event returns no error",
			rule:          "FREQ=MINUTELY;INTERVAL=59;COUNT=1",
			expectedError: "",
		},
		{
			name:          "single occurrence rrule returns no error",
			rule:          "FREQ=MINUTELY;COUNT=1",
			expectedError: "",
		},
		{
			name:          "two times minutely occurrence rrule returns error",
			rule:          "FREQ=MINUTELY;COUNT=2",
			expectedError: "interval must be at least 60 minutes for minutely frequency",
		},
		{
			name:          "hourly frequency returns no error",
			rule:          "FREQ=HOURLY;INTERVAL=1",
			expectedError: "",
		},
		{
			name:          "daily frequency returns no error",
			rule:          "FREQ=DAILY;INTERVAL=1",
			expectedError: "",
		},
		{
			name:          "frequency greater than hourly in minutes no error",
			rule:          "FREQ=MINUTELY;INTERVAL=61;COUNT=10",
			expectedError: "",
		},
		{
			name:          "frequency greater than hourly in seconds no error",
			rule:          "FREQ=SECONDLY;INTERVAL=3600;COUNT=10",
			expectedError: "",
		},
		{
			name:          "frequency greater than hourly in seconds no error",
			rule:          "FREQ=SECONDLY;INTERVAL=3600",
			expectedError: "",
		},
		{
			name:          "frequency shorter than hourly in seconds returns error",
			rule:          "FREQ=SECONDLY;INTERVAL=3500;COUNT=10",
			expectedError: "interval must be at least 3600 seconds for secondly frequency",
		},
		{
			name:          "minutely with by hour returns error",
			rule:          "FREQ=MINUTELY;BYHOUR=6,8,9,10,11,12,13,14,15,16;COUNT=10",
			expectedError: "interval must be at least 60 minutes for minutely frequency",
		},
		{
			name:          "minutely with by hour returns error",
			rule:          "FREQ=MINUTELY;INTERVAL=10;BYHOUR=6,8,9,10,11,12,13,14,15,16",
			expectedError: "interval must be at least 60 minutes for minutely frequency",
		},
		{
			name:          "hourly with by second returns error",
			rule:          "FREQ=HOURLY;BYSECOND=6,8,9,10,11,12,13,14,15,16",
			expectedError: "byminute and bysecond are not supported",
		},
		{
			name:          "hourly with by minute returns no error",
			rule:          "FREQ=HOURLY;BYHOUR=6,8,9,10,11,12,13,14,15,16;BYMINUTE=6,8,9,10,11,12,13,14,15,16,59;COUNT=10",
			expectedError: "byminute and bysecond are not supported",
		},
		{
			name:          "single by minute is supported",
			rule:          "FREQ=HOURLY;BYMINUTE=6;COUNT=10",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule *rrule.RRule
			var err error
			if tt.rule != "" {
				rule, err = rrule.StrToRRule(tt.rule)
				assert.NoError(t, err)
			}
			err = atLeastHourlyFreq.Validate(rule)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestAtLeastSecondTimeResolution(t *testing.T) {
	tests := []struct {
		name          string
		time          time.Time
		expectedError string
	}{
		{
			name:          "time with nanosecond returns error",
			time:          time.Date(2023, time.January, 1, 0, 0, 0, 1, time.UTC),
			expectedError: "time must be defined with 1s precision",
		},
		{
			name:          "time with second resolution returns no error",
			time:          time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedError: "",
		},
		{
			name:          "time with millisecond resolution returns error",
			time:          time.Date(2023, time.January, 1, 0, 0, 0, 1000000, time.UTC),
			expectedError: "time must be defined with 1s precision",
		},
		{
			name:          "time with microsecond resolution returns error",
			time:          time.Date(2023, time.January, 1, 0, 0, 0, 526, time.UTC),
			expectedError: "time must be defined with 1s precision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := secondTimePrecision.Validate(tt.time)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
