package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/validation"
)

func TestReplayStructDatesValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		replay    Replay
		isValid   bool
		ErrorCode validation.ErrorCode
	}{
		{
			name: "correct struct",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid: true,
		},
		{
			name: "missing slo",
			replay: Replay{
				Project: "project",
				Slo:     "",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeRequired,
		},
		{
			name: "missing project",
			replay: Replay{
				Project: "",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeRequired,
		},
		{
			name: "missing duration unit",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeRequired,
		},
		{
			name: "missing duration value",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit: "Day",
				},
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeGreaterThan,
		},
		{
			name: "invalid duration unit",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Test",
					Value: 30,
				},
			},
			isValid:   false,
			ErrorCode: replayDurationUnitValidationErrorCode,
		},
		{
			name: "invalid duration value",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: -30,
				},
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeGreaterThan,
		},
		{
			name: "maximum duration exceeded",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 31,
				},
			},
			isValid:   false,
			ErrorCode: replayDurationValidationErrorCode,
		},
		{
			name: "missing duration",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
			},
			isValid:   false,
			ErrorCode: validation.ErrorCodeRequired,
		},
		{
			name: "correct struct start date",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				TimeRange: ReplayTimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid: true,
		},
		{
			name: "only one of duration or start date can be set",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 30,
				},
				TimeRange: ReplayTimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid:   false,
			ErrorCode: replayDurationAndStartDateValidationError,
		},
		{
			name: "use start date without duration",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "",
					Value: 0,
				},
				TimeRange: ReplayTimeRange{
					StartDate: time.Now().Add(-time.Hour * 24),
				},
			},
			isValid: true,
		},
		{
			name: "only one of duration",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 30,
				},
				TimeRange: ReplayTimeRange{
					StartDate: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
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
				require.IsType(t, &validation.ValidatorError{}, err)
				assert.True(t, validation.HasErrorCode(err, tc.ErrorCode))
			}
		})
	}
}

func TestParseJSONToReplayStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputJSON string
		want      Replay
		wantErr   bool
	}{
		{
			name:      "pass valid json",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Day", "value": 20}}`,
			want: Replay{
				Project: "default",
				Slo:     "annotation-test",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: 20,
				},
			},
			wantErr: false,
		},
		{
			name:      "pass invalid json",
			inputJSON: `}`,
			want:      Replay{},
			wantErr:   true,
		},
		{
			name:      "pass invalid values",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Days", "value": 20}}`,
			want:      Replay{},
			wantErr:   true,
		},
		{
			name:      "pass empty object",
			inputJSON: `{}`,
			want:      Replay{},
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
			if err := ValidateReplayDurationUnit(tc.unit); err != tc.wantErr {
				t.Errorf("ValidateReplayDurationUnit() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConvertReplayDurationToDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		duration     ReplayDuration
		wantErr      error
		wantDuration time.Duration
	}{
		{
			name: "30 minutes",
			duration: ReplayDuration{
				Unit:  DurationUnitMinute,
				Value: 30,
			},
			wantErr:      nil,
			wantDuration: 30 * time.Minute,
		},
		{
			name: "15 days",
			duration: ReplayDuration{
				Unit:  DurationUnitDay,
				Value: 15,
			},
			wantErr:      nil,
			wantDuration: 24 * time.Hour * 15,
		},
		{
			name: "5 hours",
			duration: ReplayDuration{
				Unit:  DurationUnitHour,
				Value: 5,
			},
			wantErr:      nil,
			wantDuration: 5 * time.Hour,
		},
		{
			name: "invalid time unit",
			duration: ReplayDuration{
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
