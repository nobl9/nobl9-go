package v1alpha

import (
	"strings"
	"testing"
	"time"

	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestReplayStructDatesValidation(t *testing.T) {
	t.Parallel()

	validate := v.New()
	validate.RegisterStructValidation(replayStructDatesValidation, Replay{})

	tests := []struct {
		name      string
		replay    Replay
		isValid   bool
		errorTags map[string][]string
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
			isValid: false,
			errorTags: map[string][]string{
				"Slo": {"required"},
			},
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
			isValid: false,
			errorTags: map[string][]string{
				"Project": {"required"},
			},
		},
		{
			name: "missing durationUnit",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Value: 30,
				},
			},
			isValid: false,
			errorTags: map[string][]string{
				"Unit": {"required", "invalidDurationUnit"},
			},
		},
		{
			name: "missing durationValue",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit: "Day",
				},
			},
			isValid: false,
			errorTags: map[string][]string{
				"Value": {"required"},
			},
		},
		{
			name: "invalid durationUnit",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Test",
					Value: 30,
				},
			},
			isValid: false,
			errorTags: map[string][]string{
				"Unit": {"invalidDurationUnit"},
			},
		},
		{
			name: "invalid durationValue",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
				Duration: ReplayDuration{
					Unit:  "Day",
					Value: -30,
				},
			},
			isValid: false,
			errorTags: map[string][]string{
				"Value": {"gte"},
			},
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
			isValid: false,
			errorTags: map[string][]string{
				"Value": {"maximumDurationExceeded"},
			},
		},
		{
			name: "missing duration",
			replay: Replay{
				Project: "project",
				Slo:     "slo",
			},
			isValid: false,
			errorTags: map[string][]string{
				"Unit":  {"required", "invalidDurationUnit"},
				"Value": {"required"},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Struct(tc.replay)
			if tc.isValid {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)

				// check all error tags
				tags := map[string][]string{}
				errors := err.(v.ValidationErrors)
				for i := range errors {
					fe := errors[i]

					tags[fe.StructField()] = append(tags[fe.StructField()], fe.Tag())
				}

				assert.Equal(t, tc.errorTags, tags)
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

			assert.Equal(t, tc.wantErr, err != nil)
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
