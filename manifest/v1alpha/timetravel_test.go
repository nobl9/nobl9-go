//go:build unit_test

package v1alpha

import (
	"strings"
	"testing"
	"time"

	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestTimeTravelStructDatesValidation(t *testing.T) {
	t.Parallel()

	validate := v.New()
	validate.RegisterStructValidation(timeTravelStructDatesValidation, TimeTravel{})

	tests := []struct {
		name       string
		timeTravel TimeTravel
		isValid    bool
		errorTags  map[string][]string
	}{
		{
			name: "correct struct",
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
					Unit:  "Day",
					Value: 30,
				},
			},
			isValid: true,
		},
		{
			name: "missing slo",
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
				Project: "project",
				Slo:     "slo",
				Duration: TimeTravelDuration{
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
			timeTravel: TimeTravel{
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

			err := validate.Struct(tc.timeTravel)
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

func TestParseJSONToTimeTravelStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputJSON string
		want      TimeTravel
		wantErr   bool
	}{
		{
			name:      "pass valid json",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Day", "value": 20}}`,
			want: TimeTravel{
				Project: "default",
				Slo:     "annotation-test",
				Duration: TimeTravelDuration{
					Unit:  "Day",
					Value: 20,
				},
			},
			wantErr: false,
		},
		{
			name:      "pass invalid json",
			inputJSON: `}`,
			want:      TimeTravel{},
			wantErr:   true,
		},
		{
			name:      "pass invalid values",
			inputJSON: `{"project": "default","slo": "annotation-test", "duration": {"unit": "Days", "value": 20}}`,
			want:      TimeTravel{},
			wantErr:   true,
		},
		{
			name:      "pass empty object",
			inputJSON: `{}`,
			want:      TimeTravel{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tc.inputJSON)
			got, err := ParseJSONToTimeTravelStruct(reader)

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
			wantErr: ErrInvalidTimeTravelDurationUnit,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateTimeTravelDurationUnit(tc.unit); err != tc.wantErr {
				t.Errorf("ValidateTimeTravelDurationUnit() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConvertTimeTravelDurationToDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		duration     TimeTravelDuration
		wantErr      error
		wantDuration time.Duration
	}{
		{
			name: "30 minutes",
			duration: TimeTravelDuration{
				Unit:  DurationUnitMinute,
				Value: 30,
			},
			wantErr:      nil,
			wantDuration: 30 * time.Minute,
		},
		{
			name: "15 days",
			duration: TimeTravelDuration{
				Unit:  DurationUnitDay,
				Value: 15,
			},
			wantErr:      nil,
			wantDuration: 24 * time.Hour * 15,
		},
		{
			name: "5 hours",
			duration: TimeTravelDuration{
				Unit:  DurationUnitHour,
				Value: 5,
			},
			wantErr:      nil,
			wantDuration: 5 * time.Hour,
		},
		{
			name: "invalid time unit",
			duration: TimeTravelDuration{
				Unit:  "TEST",
				Value: 30,
			},
			wantErr:      ErrInvalidTimeTravelDurationUnit,
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
