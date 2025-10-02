package v1

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

func TestRunRequestDatesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		replay    RunRequest
		isValid   bool
		ErrorCode govy.ErrorCode
	}{
		{
			name: "correct struct",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid: true,
		},
		{
			name: "missing slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing project",
			replay: RunRequest{
				Project: "",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeRequired,
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
			ErrorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing duration value",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit: "Day",
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "invalid duration unit",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Test",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: durationUnitValidationErrorCode,
		},
		{
			name: "invalid duration value",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: -30,
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeGreaterThan,
		},
		{
			name: "maximum duration exceeded",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 31,
				},
			},
			isValid:   false,
			ErrorCode: durationValidationErrorCode,
		},
		{
			name: "missing duration",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeRequired,
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
					Unit:  "Day",
					Value: 30,
				},
				TimeRange: TimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid:   false,
			ErrorCode: durationAndStartDateValidationError,
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
			ErrorCode: startDateInTheFutureValidationError,
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
			name: "only one of duration",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
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
					Unit:  "Day",
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "project",
					SLO:     "",
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeRequired,
		},
		{
			name: "source project is required",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "",
					SLO:     "slo",
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeRequired,
		},
		{
			name: "missing objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project: "project",
					SLO:     "slo",
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeSliceMinLength,
		},
		{
			name: "empty objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
					Value: 30,
				},
				SourceSLO: &SourceSLO{
					Project:       "project",
					SLO:           "slo",
					ObjectivesMap: []SourceSLOItem{},
				},
			},
			isValid:   false,
			ErrorCode: rules.ErrorCodeSliceMinLength,
		},
		{
			name: "not empty objectives map when replaying source slo",
			replay: RunRequest{
				Project: "project",
				SLO:     "slo",
				Duration: Duration{
					Unit:  "Day",
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
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.replay.Validate()
			if tc.isValid {
				assert.Nil(t, err)
			} else {
				require.Error(t, err)
				require.IsType(t, &govy.ValidatorError{}, err)
				assert.True(t, govy.HasErrorCode(err, tc.ErrorCode))
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
			name:      "pass valid json",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Day", "value": 20}}`,
			want: RunRequest{
				Project: "default",
				SLO:     "annotation-test",
				Duration: Duration{
					Unit:  "Day",
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
			name:      "pass empty object",
			inputJSON: `{}`,
			want:      RunRequest{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tc.inputJSON)
			got, err := ParseJSONToReplayStruct(reader)

			if tc.wantErr {
				assert.NotEmpty(t, err)
			} else {
				assert.Empty(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCheckPeriodUnit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		unit    string
		wantErr error
	}{
		{
			name:    "Proper duration unit",
			unit:    DurationUnitDay,
			wantErr: nil,
		},
		{
			name:    "Invalid duration unit",
			unit:    "Days",
			wantErr: ErrInvalidReplayDurationUnit,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateDurationUnit(tc.unit); err != tc.wantErr {
				t.Errorf("ValidateReplayDurationUnit() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConvertDurationToTimeDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		duration     Duration
		wantErr      error
		wantDuration time.Duration
	}{
		{
			name: "30 minutes",
			duration: Duration{
				Unit:  DurationUnitMinute,
				Value: 30,
			},
			wantErr:      nil,
			wantDuration: 30 * time.Minute,
		},
		{
			name: "15 days",
			duration: Duration{
				Unit:  DurationUnitDay,
				Value: 15,
			},
			wantErr:      nil,
			wantDuration: 24 * time.Hour * 15,
		},
		{
			name: "5 hours",
			duration: Duration{
				Unit:  DurationUnitHour,
				Value: 5,
			},
			wantErr:      nil,
			wantDuration: 5 * time.Hour,
		},
		{
			name: "invalid time unit",
			duration: Duration{
				Unit:  "TEST",
				Value: 30,
			},
			wantErr:      ErrInvalidReplayDurationUnit,
			wantDuration: 0,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			duration, err := tc.duration.Duration()

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantDuration, duration)
		})
	}
}
