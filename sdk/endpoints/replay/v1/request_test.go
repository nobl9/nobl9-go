package v1

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"
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

func TestParseJSONToRunRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputJSON string
		want      RunRequest
		wantErr   bool
	}{
		{
			name: "pass valid json",
			inputJSON: `{
				"project": "default",
				"slo": "annotation-test",
				"replayType": "recalculation",
				"duration": {
					"unit": "Day",
					"value": 20
				}
			}`,
			want: RunRequest{
				Project:    "default",
				SLO:        "annotation-test",
				ReplayType: ReplayTypeRecalculation,
				Duration: Duration{
					Unit:  DurationUnitDay,
					Value: 20,
				},
			},
			wantErr: false,
		},
		{
			name:      "pass invalid json",
			inputJSON: `}`,
			want:      RunRequest{},
			wantErr:   true,
		},
		{
			name:      "pass invalid values",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Days", "value": 20}}`,
			want:      RunRequest{},
			wantErr:   true,
		},
		{
			name: "pass invalid replay type",
			inputJSON: `{
				"project": "default",
				"slo": "annotation-test",
				"replayType": "unsupported",
				"duration": {
					"unit": "Day",
					"value": 20
				}
			}`,
			want:    RunRequest{},
			wantErr: true,
		},
		{
			name:      "pass empty object",
			inputJSON: `{}`,
			want:      RunRequest{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tt.inputJSON)
			got, err := ParseJSONToReplayStruct(reader)

			if tt.wantErr {
				assert.NotEmpty(t, err)
			} else {
				assert.Empty(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
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

func TestGetAvailabilityRequestQueryValues(t *testing.T) {
	t.Parallel()

	values := GetAvailabilityRequest{
		Project:           "request-project",
		DataSourceProject: "source-project",
		DataSource:        "datadog",
		DataSourceKind:    "Direct",
		SLOName:           "latency-slo",
		Type:              ReplayTypeRecalculation,
		DurationUnit:      DurationUnitHour,
		DurationValue:     1,
	}.queryValues()

	assert.Equal(t, url.Values{
		"dataSourceProject": {"source-project"},
		"dataSource":        {"datadog"},
		"dataSourceKind":    {"Direct"},
		"sloName":           {"latency-slo"},
		"type":              {"recalculation"},
		"durationUnit":      {"Hour"},
		"durationValue":     {"1"},
	}, values)
}

func TestDuration_Duration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		duration     Duration
		wantDuration time.Duration
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
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			duration := tc.duration.Duration()
			assert.Equal(t, tc.wantDuration, duration)
		})
	}
}
