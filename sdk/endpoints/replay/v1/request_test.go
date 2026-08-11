package v1

import (
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
)

func TestRunRequestDatesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		replay    RunRequest
		isValid   bool
		errorCode govy.ErrorCode
	}{
		{
			name: "correct struct",
			replay: RunRequest{
				Project:    "project",
				SLO:        "slo",
				ReplayType: ReplayTypeRecalculation,
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
			},
			isValid: true,
		},
		{
			name: "invalid replay type",
			replay: RunRequest{
				Project:    "project",
				SLO:        "slo",
				ReplayType: ReplayType("unsupported"),
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeOneOf,
		},
		{
			name: "missing slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing project",
			replay: RunRequest{
				Project: "",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing duration unit",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Value: 30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing duration value",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit: DurationUnitDay,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "invalid duration unit",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnit("Test"),
					Value: 30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeOneOf,
		},
		{
			name: "invalid duration value",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: -30,
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "duration over 30 days",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 31,
				},
			},
			isValid: true,
		},
		{
			name: "missing duration",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "correct struct start date",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				TimeRange: TimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid: true,
		},
		{
			name: "only one of duration or start date can be set",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				TimeRange: TimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid:   false,
			errorCode: durationAndStartDateValidationError,
		},
		{
			name: "start date cannot be in the future",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				TimeRange: TimeRange{
					StartDate: time.Now().Add(time.Minute * 1),
				},
			},
			isValid:   false,
			errorCode: startDateInTheFutureValidationError,
		},
		{
			name: "use start date without duration",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "",
					Value: 0,
				},
				TimeRange: TimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid: true,
		},
		{
			name: "partial duration with start date",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Value: 30,
				},
				TimeRange: TimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "only one of duration",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				TimeRange: TimeRange{
					StartDate: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			isValid: true,
		},
		{
			name: "source slo is required",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "project",
					SLO:     "",
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "source project is required",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "",
					SLO:     "slo",
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "project",
					SLO:     "slo",
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeSliceMinLength,
		},
		{
			name: "empty objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project:       "project",
					SLO:           "slo",
					ObjectivesMap: []SourceSLOItem{},
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeSliceMinLength,
		},
		{
			name: "source objective is required",
			replay: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: Duration{Unit: DurationUnitDay, Value: 30},
				SourceSLO: &SourceSLO{
					Project:       "project",
					SLO:           "slo",
					ObjectivesMap: []SourceSLOItem{{Target: "objective-1"}},
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "target objective is required",
			replay: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: Duration{Unit: DurationUnitDay, Value: 30},
				SourceSLO: &SourceSLO{
					Project:       "project",
					SLO:           "slo",
					ObjectivesMap: []SourceSLOItem{{Source: "objective-1"}},
				},
			},
			isValid:   false,
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "not empty objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "project",
					SLO:     "slo",
					ObjectivesMap: []SourceSLOItem{
						{
							Source: "objective-1",
							Target: "objective-1",
						},
					},
				},
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.replay.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				require.IsType(t, &govy.ValidatorError{}, err)
				assert.True(t, govy.HasErrorCode(err, tt.errorCode))
			}
		})
	}
}

func TestRunRequestMarshalSourceSLO(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(RunRequest{
		Project:  "target-project",
		SLO:      "target-slo",
		Duration: Duration{Unit: DurationUnitHour, Value: 1},
		SourceSLO: &SourceSLO{
			Project: "source-project",
			SLO:     "source-slo",
			ObjectivesMap: []SourceSLOItem{
				{Source: "source-objective", Target: "target-objective"},
			},
		},
	})

	require.NoError(t, err)
	assert.JSONEq(t, `{
		"sourceSlo": {
			"slo": "source-slo",
			"project": "source-project",
			"objectivesMap": [
				{"source": "source-objective", "target": "target-objective"}
			]
		},
		"project": "target-project",
		"slo": "target-slo",
		"duration": {"unit": "Hour", "value": 1}
	}`, string(data))
}

func TestRunRequestMarshalTimeRange(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2026, time.August, 11, 12, 0, 0, 0, time.UTC)
	request := RunRequest{
		Project: "project",
		SLO:     "slo",
		TimeRange: TimeRange{
			StartDate: startDate,
			EndDate:   startDate.Add(time.Hour),
		},
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"project": "project",
		"slo": "slo",
		"timeRange": {
			"startDate": "2026-08-11T12:00:00Z",
			"endDate": "2026-08-11T13:00:00Z"
		}
	}`, string(data))

	request.TimeRange.EndDate = time.Time{}
	data, err = json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"project": "project",
		"slo": "slo",
		"timeRange": {
			"startDate": "2026-08-11T12:00:00Z"
		}
	}`, string(data))
}

func TestGetAvailabilityRequestValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        GetAvailabilityRequest
		isValid        bool
		errorCode      govy.ErrorCode
		expectedErrors []govytest.ExpectedRuleError
	}{
		{
			name: "existing slo with default project",
			request: GetAvailabilityRequest{
				SLOName: "slo",
			},
			isValid: true,
		},
		{
			name: "data source with duration",
			request: GetAvailabilityRequest{
				Project:           "project",
				DataSourceProject: "data-source-project",
				DataSource:        "data-source",
				DataSourceKind:    "Direct",
				Type:              ReplayTypeReimportAndRecalculation,
				DurationUnit:      DurationUnitHour,
				DurationValue:     1,
			},
			isValid: true,
		},
		{
			name: "missing data source selection",
			request: GetAvailabilityRequest{
				Project: "project",
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{PropertyPath: "dataSourceProject", Code: rules.ErrorCodeRequired},
				{PropertyPath: "dataSource", Code: rules.ErrorCodeRequired},
				{PropertyPath: "dataSourceKind", Code: rules.ErrorCodeRequired},
			},
		},
		{
			name: "invalid replay type",
			request: GetAvailabilityRequest{
				Project: "project",
				SLOName: "slo",
				Type:    ReplayType("unsupported"),
			},
			errorCode: rules.ErrorCodeOneOf,
		},
		{
			name: "non-data-source manifest kind is handled by the server",
			request: GetAvailabilityRequest{
				DataSourceProject: "data-source-project",
				DataSource:        "data-source",
				DataSourceKind:    "SLO",
			},
			isValid: true,
		},
		{
			name: "invalid data source kind",
			request: GetAvailabilityRequest{
				DataSourceProject: "data-source-project",
				DataSource:        "data-source",
				DataSourceKind:    "not-a-kind",
			},
			errorCode: rules.ErrorCodeOneOf,
		},
		{
			name: "mixed SLO and data source selectors",
			request: GetAvailabilityRequest{
				SLOName:           "slo",
				DataSourceProject: "data-source-project",
				DataSource:        "data-source",
				DataSourceKind:    "Direct",
			},
			errorCode: rules.ErrorCodeMutuallyExclusive,
		},
		{
			name: "duration unit without value",
			request: GetAvailabilityRequest{
				Project:      "project",
				SLOName:      "slo",
				DurationUnit: DurationUnitHour,
			},
			errorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "duration value without unit",
			request: GetAvailabilityRequest{
				Project:       "project",
				SLOName:       "slo",
				DurationValue: 1,
			},
			errorCode: rules.ErrorCodeRequired,
		},
		{
			name: "negative duration",
			request: GetAvailabilityRequest{
				Project:       "project",
				SLOName:       "slo",
				DurationUnit:  DurationUnitHour,
				DurationValue: -1,
			},
			errorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "invalid duration unit",
			request: GetAvailabilityRequest{
				Project:       "project",
				SLOName:       "slo",
				DurationUnit:  DurationUnit("Hours"),
				DurationValue: 1,
			},
			errorCode: rules.ErrorCodeOneOf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.request.Validate()
			if tt.isValid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.IsType(t, &govy.ValidatorError{}, err)
			if len(tt.expectedErrors) > 0 {
				govytest.AssertError(t, err, tt.expectedErrors...)
				return
			}
			assert.True(t, govy.HasErrorCode(err, tt.errorCode))
		})
	}
}

func TestDeleteRequestValidation(t *testing.T) {
	t.Parallel()

	require.NoError(t, DeleteRequest{All: true}.Validate())
	require.NoError(t, DeleteRequest{Project: "BadProject", SLO: "team/slo", All: true}.Validate())
	require.NoError(t, DeleteRequest{Project: "project", SLO: "slo"}.Validate())
	require.Error(t, DeleteRequest{}.Validate())
	require.Error(t, DeleteRequest{Project: "project"}.Validate())
	require.Error(t, DeleteRequest{SLO: "slo"}.Validate())
}

func TestCancelRequestValidation(t *testing.T) {
	t.Parallel()

	require.NoError(t, CancelRequest{Project: "project", SLO: "slo"}.Validate())
	require.Error(t, CancelRequest{}.Validate())
	require.Error(t, CancelRequest{Project: "project"}.Validate())
	require.Error(t, CancelRequest{SLO: "slo"}.Validate())
}

func TestGetStatusRequestValidation(t *testing.T) {
	t.Parallel()

	require.NoError(t, GetStatusRequest{SLO: "slo"}.Validate())
	require.NoError(t, GetStatusRequest{Project: "project", SLO: "slo"}.Validate())
	require.Error(t, GetStatusRequest{}.Validate())
	require.Error(t, GetStatusRequest{Project: "project"}.Validate())
}

func TestReplaySelectorValidation(t *testing.T) {
	t.Parallel()

	validDuration := Duration{Unit: DurationUnitHour, Value: 1}
	tooLong := strings.Repeat("a", validationV1Alpha.NameMaximumLength+1)
	tests := []struct {
		name         string
		propertyPath string
		request      interface{ Validate() error }
	}{
		{
			name:         "uppercase run project",
			propertyPath: "project",
			request:      RunRequest{Project: "Project", SLO: "slo", Duration: validDuration},
		},
		{
			name:         "slash in run SLO",
			propertyPath: "slo",
			request:      RunRequest{Project: "project", SLO: "team/slo", Duration: validDuration},
		},
		{
			name:         "source SLO project is too long",
			propertyPath: "sourceSLO.project",
			request: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: validDuration,
				SourceSLO: &SourceSLO{
					Project:       tooLong,
					SLO:           "source-slo",
					ObjectivesMap: []SourceSLOItem{{Source: "source", Target: "target"}},
				},
			},
		},
		{
			name:         "slash in source SLO",
			propertyPath: "sourceSLO.slo",
			request: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: validDuration,
				SourceSLO: &SourceSLO{
					Project:       "source-project",
					SLO:           "team/source-slo",
					ObjectivesMap: []SourceSLOItem{{Source: "source", Target: "target"}},
				},
			},
		},
		{
			name:         "uppercase source objective",
			propertyPath: "sourceSLO.objectivesMap[0].source",
			request: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: validDuration,
				SourceSLO: &SourceSLO{
					Project:       "source-project",
					SLO:           "source-slo",
					ObjectivesMap: []SourceSLOItem{{Source: "Source", Target: "target"}},
				},
			},
		},
		{
			name:         "slash in target objective",
			propertyPath: "sourceSLO.objectivesMap[0].target",
			request: RunRequest{
				Project:  "project",
				SLO:      "slo",
				Duration: validDuration,
				SourceSLO: &SourceSLO{
					Project:       "source-project",
					SLO:           "source-slo",
					ObjectivesMap: []SourceSLOItem{{Source: "source", Target: "team/target"}},
				},
			},
		},
		{
			name:         "slash in delete project",
			propertyPath: "project",
			request:      DeleteRequest{Project: "team/project", SLO: "slo"},
		},
		{
			name:         "uppercase delete SLO",
			propertyPath: "slo",
			request:      DeleteRequest{Project: "project", SLO: "SLO"},
		},
		{
			name:         "slash in cancel project",
			propertyPath: "project",
			request:      CancelRequest{Project: "team/project", SLO: "slo"},
		},
		{
			name:         "uppercase cancel SLO",
			propertyPath: "slo",
			request:      CancelRequest{Project: "project", SLO: "SLO"},
		},
		{
			name:         "slash in status project",
			propertyPath: "project",
			request:      GetStatusRequest{Project: "team/project", SLO: "slo"},
		},
		{
			name:         "uppercase status SLO",
			propertyPath: "slo",
			request:      GetStatusRequest{Project: "project", SLO: "SLO"},
		},
		{
			name:         "slash in availability project",
			propertyPath: "project",
			request:      GetAvailabilityRequest{Project: "team/project", SLOName: "slo"},
		},
		{
			name:         "availability SLO is too long",
			propertyPath: "sloName",
			request:      GetAvailabilityRequest{SLOName: tooLong},
		},
		{
			name:         "uppercase availability data source",
			propertyPath: "dataSource",
			request: GetAvailabilityRequest{
				DataSourceProject: "data-source-project",
				DataSource:        "DataSource",
				DataSourceKind:    "Direct",
			},
		},
		{
			name:         "slash in availability data source project",
			propertyPath: "dataSourceProject",
			request: GetAvailabilityRequest{
				DataSourceProject: "team/project",
				DataSource:        "data-source",
				DataSourceKind:    "Direct",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.request.Validate()
			require.Error(t, err)
			govytest.AssertError(t, err, govytest.ExpectedRuleError{
				PropertyPath: tt.propertyPath,
				Code:         validationV1Alpha.ErrorCodeStringName,
			})
		})
	}
}

func TestGetAvailabilityRequestQueryValues(t *testing.T) {
	t.Parallel()

	dataSourceValues := GetAvailabilityRequest{
		Project:           "request-project",
		DataSourceProject: "source-project",
		DataSource:        "datadog",
		DataSourceKind:    "Direct",
		Type:              ReplayTypeRecalculation,
		DurationUnit:      DurationUnitHour,
		DurationValue:     1,
	}.queryValues()

	assert.Equal(t, url.Values{
		"dataSourceProject": {"source-project"},
		"dataSource":        {"datadog"},
		"dataSourceKind":    {"Direct"},
		"type":              {"recalculation"},
		"durationUnit":      {"Hour"},
		"durationValue":     {"1"},
	}, dataSourceValues)

	sloValues := GetAvailabilityRequest{SLOName: "latency-slo"}.queryValues()
	assert.Equal(t, url.Values{"sloName": {"latency-slo"}}, sloValues)
}

func TestDuration_Duration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		duration     Duration
		wantDuration time.Duration
		wantErr      error
	}{
		{
			name: "30 minutes",
			duration: Duration{
				Unit:  DurationUnitMinute,
				Value: 30,
			},
			wantDuration: 30 * time.Minute,
		},
		{
			name: "15 days",
			duration: Duration{
				Unit:  DurationUnitDay,
				Value: 15,
			},
			wantDuration: 24 * time.Hour * 15,
		},
		{
			name: "5 hours",
			duration: Duration{
				Unit:  DurationUnitHour,
				Value: 5,
			},
			wantDuration: 5 * time.Hour,
		},
		{
			name: "invalid time unit",
			duration: Duration{
				Unit:  DurationUnit("TEST"),
				Value: 30,
			},
			wantDuration: 0,
			wantErr:      ErrInvalidDurationUnit,
		},
		{
			name: "zero value",
			duration: Duration{
				Unit: DurationUnitHour,
			},
			wantErr: ErrInvalidDurationValue,
		},
		{
			name: "negative value",
			duration: Duration{
				Unit:  DurationUnitHour,
				Value: -1,
			},
			wantErr: ErrInvalidDurationValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			duration, err := tt.duration.Duration()
			assert.Equal(t, tt.wantDuration, duration)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			if errors.Is(tt.wantErr, ErrInvalidDurationUnit) {
				assert.Contains(t, err.Error(), "Minute, Hour, Day")
			}
		})
	}
}

func TestDuration_DurationOverflow(t *testing.T) {
	t.Parallel()

	for unit, multiplier := range map[DurationUnit]time.Duration{
		DurationUnitMinute: time.Minute,
		DurationUnitHour:   time.Hour,
		DurationUnitDay:    24 * time.Hour,
	} {
		t.Run(unit.String(), func(t *testing.T) {
			t.Parallel()

			maxValue := math.MaxInt64 / int64(multiplier)
			if maxValue >= int64(math.MaxInt) {
				t.Skip("time.Duration cannot overflow with an int-sized value on this platform")
			}

			duration, err := (Duration{Unit: unit, Value: int(maxValue)}).Duration()
			require.NoError(t, err)
			assert.Equal(t, time.Duration(maxValue)*multiplier, duration)

			duration, err = (Duration{Unit: unit, Value: int(maxValue + 1)}).Duration()
			require.ErrorIs(t, err, ErrDurationOverflow)
			assert.Zero(t, duration)
		})
	}
}
