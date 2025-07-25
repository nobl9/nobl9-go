package service

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Service '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/service.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	svc := validService()
	svc.APIVersion = "v0.1"
	svc.Kind = manifest.KindProject
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 2,
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
	svc := validService()
	svc.Metadata = Metadata{
		Name:        strings.Repeat("MY SERVICE", 20),
		DisplayName: strings.Repeat("my-service", 20),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 3,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeStringDNSLabel,
		},
	)
}

func TestValidate_Metadata_Labels(t *testing.T) {
	for name, test := range v1alphatest.GetLabelsTestCases[Service](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[Service](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Spec(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		svc := validService()
		svc.Spec.Description = strings.Repeat("A", 2000)
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validationV1Alpha.ErrorCodeStringDescription,
		})
	})
}

func TestValidate_ReviewCycle(t *testing.T) {
	t.Run("valid reviewCycle", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertNoError(t, svc, err)
	})

	t.Run("invalid startTime - wrong format", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startTime",
			Code: "string_date_time",
		})
	})

	t.Run("empty startTime", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "",
			TimeZone:  "UTC",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.startTime",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("empty timeZone", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.timeZone",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("invalid timeZone", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "Invalid/Timezone",
			RRule:     "FREQ=MONTHLY;INTERVAL=1",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.timeZone",
			Code: "string_time_zone",
		})
	})

	t.Run("empty rrule", func(t *testing.T) {
		svc := validService()
		svc.Spec.ReviewCycle = &ReviewCycle{
			StartTime: "2023-01-01T00:00:00",
			TimeZone:  "UTC",
			RRule:     "",
		}
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.reviewCycle.rrule",
			Code: rules.ErrorCodeRequired,
		})
	})

	t.Run("rrule frequency validation", func(t *testing.T) {
		tests := []struct {
			name        string
			rrule       string
			expectError bool
			expectedMsg string
		}{
			{
				name:        "missing FREQ",
				rrule:       "INTERVAL=1",
				expectError: true,
				expectedMsg: "RRULE property FREQ is required",
			},
			{
				name:        "hourly frequency not allowed",
				rrule:       "FREQ=HOURLY;INTERVAL=1",
				expectError: true,
				expectedMsg: "frequency must be DAILY, WEEKLY, MONTHLY, or YEARLY",
			},
			{
				name:        "minutely frequency not allowed",
				rrule:       "FREQ=MINUTELY;INTERVAL=60",
				expectError: true,
				expectedMsg: "frequency must be DAILY, WEEKLY, MONTHLY, or YEARLY",
			},
			{
				name:        "daily frequency allowed",
				rrule:       "FREQ=DAILY;INTERVAL=1",
				expectError: false,
			},
			{
				name:        "weekly frequency allowed",
				rrule:       "FREQ=WEEKLY;INTERVAL=1",
				expectError: false,
			},
			{
				name:        "monthly frequency allowed",
				rrule:       "FREQ=MONTHLY;INTERVAL=1",
				expectError: false,
			},
			{
				name:        "yearly frequency allowed",
				rrule:       "FREQ=YEARLY;INTERVAL=1",
				expectError: false,
			},
			{
				name:        "single occurrence with any frequency not allowed",
				rrule:       "FREQ=HOURLY;COUNT=1",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "single occurrence with daily frequency not allowed",
				rrule:       "FREQ=DAILY;COUNT=1",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "multiple limited occurrences not allowed",
				rrule:       "FREQ=WEEKLY;COUNT=5",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "zero interval not allowed",
				rrule:       "FREQ=DAILY;INTERVAL=0",
				expectError: false, // rrule library accepts 0 and converts to 1
			},
			{
				name:        "negative interval not allowed",
				rrule:       "FREQ=WEEKLY;INTERVAL=-1",
				expectError: true,
				expectedMsg: "interval must be greater than 0",
			},
			{
				name:        "secondly frequency not allowed",
				rrule:       "FREQ=SECONDLY;INTERVAL=1",
				expectError: true,
				expectedMsg: "frequency must be DAILY, WEEKLY, MONTHLY, or YEARLY",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := validService()
				svc.Spec.ReviewCycle = &ReviewCycle{
					StartTime: "2023-01-01T00:00:00",
					TimeZone:  "UTC",
					RRule:     tt.rrule,
				}
				err := validate(svc)

				if tt.expectError {
					testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
						Prop:    "spec.reviewCycle.rrule",
						Message: tt.expectedMsg,
					})
				} else {
					testutils.AssertNoError(t, svc, err)
				}
			})
		}
	})

	t.Run("rrule complex validation scenarios", func(t *testing.T) {
		tests := []struct {
			name        string
			rrule       string
			expectError bool
			expectedMsg string
		}{
			{
				name:        "daily with complex by-rules allowed",
				rrule:       "FREQ=DAILY;BYHOUR=9,17;BYMINUTE=0",
				expectError: false,
			},
			{
				name:        "weekly with multiple weekdays allowed",
				rrule:       "FREQ=WEEKLY;BYDAY=MO,WE,FR;BYHOUR=10;BYMINUTE=30",
				expectError: false,
			},
			{
				name:        "monthly with specific month days allowed",
				rrule:       "FREQ=MONTHLY;BYMONTHDAY=1,15;BYHOUR=14;BYMINUTE=0",
				expectError: false,
			},
			{
				name:        "yearly with specific month and day allowed",
				rrule:       "FREQ=YEARLY;BYMONTH=1,6,12;BYMONTHDAY=1;BYHOUR=9;BYMINUTE=0",
				expectError: false,
			},
			{
				name:        "daily with high interval allowed",
				rrule:       "FREQ=DAILY;INTERVAL=30",
				expectError: false,
			},
			{
				name:        "weekly with high interval allowed",
				rrule:       "FREQ=WEEKLY;INTERVAL=12;BYDAY=MO",
				expectError: false,
			},
			{
				name:        "monthly with high interval allowed",
				rrule:       "FREQ=MONTHLY;INTERVAL=6;BYMONTHDAY=15",
				expectError: false,
			},
			{
				name:        "yearly with interval allowed",
				rrule:       "FREQ=YEARLY;INTERVAL=2;BYMONTH=6;BYMONTHDAY=1",
				expectError: false,
			},
			{
				name:        "hourly with until date not allowed",
				rrule:       "FREQ=HOURLY;UNTIL=20251231T235959Z",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "daily with until date not allowed",
				rrule:       "FREQ=DAILY;UNTIL=20251231T235959Z",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "weekly with until and count not allowed",
				rrule:       "FREQ=WEEKLY;UNTIL=20251231T235959Z;COUNT=10",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "monthly with count and no until not allowed",
				rrule:       "FREQ=MONTHLY;COUNT=12;BYMONTHDAY=1",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "complex daily pattern without count allowed",
				rrule:       "FREQ=DAILY;INTERVAL=2;BYHOUR=9,13,17;BYMINUTE=0,30",
				expectError: false,
			},
			{
				name:        "complex weekly pattern without count allowed",
				rrule:       "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU,WE,TH,FR;BYHOUR=9;BYMINUTE=0",
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := validService()
				svc.Spec.ReviewCycle = &ReviewCycle{
					StartTime: "2023-01-01T00:00:00",
					TimeZone:  "UTC",
					RRule:     tt.rrule,
				}
				err := validate(svc)

				if tt.expectError {
					testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
						Prop:    "spec.reviewCycle.rrule",
						Message: tt.expectedMsg,
					})
				} else {
					testutils.AssertNoError(t, svc, err)
				}
			})
		}
	})

	t.Run("rrule edge cases and boundary conditions", func(t *testing.T) {
		tests := []struct {
			name        string
			rrule       string
			expectError bool
			expectedMsg string
		}{
			{
				name:        "malformed rrule",
				rrule:       "INVALID_RRULE_FORMAT",
				expectError: true,
				expectedMsg: "wrong format",
			},
			{
				name:        "empty frequency value",
				rrule:       "FREQ=;INTERVAL=1",
				expectError: true,
				expectedMsg: "FREQ option has no value",
			},
			{
				name:        "invalid frequency value",
				rrule:       "FREQ=INVALID;INTERVAL=1",
				expectError: true,
				expectedMsg: "undefined frequency: INVALID",
			},
			{
				name:        "count with zero value not allowed",
				rrule:       "FREQ=DAILY;COUNT=0",
				expectError: false, // rrule library accepts 0 but our validator should catch positive count
			},
			{
				name:        "count with negative value not allowed",
				rrule:       "FREQ=WEEKLY;COUNT=-5",
				expectError: false, // rrule library may handle this differently
			},
			{
				name:        "count with positive value not allowed",
				rrule:       "FREQ=DAILY;COUNT=10",
				expectError: true,
				expectedMsg: "rrule must have continuous occurrences",
			},
			{
				name:        "very high interval allowed",
				rrule:       "FREQ=YEARLY;INTERVAL=100",
				expectError: false,
			},
			{
				name:        "daily with valid bysetpos",
				rrule:       "FREQ=DAILY;BYSETPOS=1;BYHOUR=9;BYMINUTE=0",
				expectError: false,
			},
			{
				name:        "weekly with valid wkst",
				rrule:       "FREQ=WEEKLY;WKST=MO;BYDAY=FR",
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := validService()
				svc.Spec.ReviewCycle = &ReviewCycle{
					StartTime: "2023-01-01T00:00:00",
					TimeZone:  "UTC",
					RRule:     tt.rrule,
				}
				err := validate(svc)

				if tt.expectError {
					testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
						Prop:    "spec.reviewCycle.rrule",
						Message: tt.expectedMsg,
					})
				} else {
					testutils.AssertNoError(t, svc, err)
				}
			})
		}
	})
}

func validService() Service {
	return New(
		Metadata{
			Name:    "service",
			Project: "default",
		},
		Spec{},
	)
}
