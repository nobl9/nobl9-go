package budgetadjustment

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

func TestValidate_Metadata(t *testing.T) {
	budgetAdjustment := BudgetAdjustment{
		Kind: manifest.KindBudgetAdjustment,
		Metadata: Metadata{
			Name:        strings.Repeat("MY BUDGET ADJUSTMENT ", 3),
			DisplayName: strings.Repeat("my-budget-adjustment-", 10),
		},
		Spec:           Spec{},
		ManifestSource: "/home/me/budget-adjustment.yaml",
	}
	err := validate(budgetAdjustment)
	assert.Error(t, err)

	expectedErrors := []testutils.ExpectedError{
		{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain + ":" + validation.ErrorCodeStringMatchRegexp,
		},
		{
			Prop: "metadata.displayName",
			Code: validation.ErrorCodeStringLength,
		},
		{
			Prop: "spec.firstEventStart",
			Code: validation.ErrorCodeRequired,
		},
		{
			Prop: "spec.duration",
			Code: validation.ErrorCodeRequired,
		},
		{
			Prop: "spec.filters.slos",
			Code: validation.ErrorCodeSliceMinLength,
		},
	}

	testutils.AssertContainsErrors(t, budgetAdjustment, err, len(expectedErrors), expectedErrors...)
}

func TestValidate_Spec(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedErrors []testutils.ExpectedError
	}{
		{
			name: "no slo filters",
			spec: Spec{
				FirstEventStart: time.Now(),
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
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeDurationFullMinutePrecision,
				},
			},
		},
		{
			name: "duration contains seconds",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeDurationFullMinutePrecision,
				},
			},
		},
		{
			name: "slo is defined without name",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeRequired,
				},
			},
		},
		{
			name: "slo is defined with invalid slo name",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeStringIsDNSSubdomain,
				},
			},
		},
		{
			name: "slo is defined without project",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeRequired,
				},
			},
		},
		{
			name: "slo is defined with invalid project name",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeStringIsDNSSubdomain,
				},
			},
		},
		{
			name: "wrong rrule format",
			spec: Spec{
				FirstEventStart: time.Now(),
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
				FirstEventStart: time.Now(),
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
			name: "proper spec",
			spec: Spec{
				FirstEventStart: time.Now(),
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
			alertMethod := BudgetAdjustment{
				Kind: manifest.KindBudgetAdjustment,
				Metadata: Metadata{
					Name: "my-budget-adjustment",
				},
				Spec:           test.spec,
				ManifestSource: "/home/me/budget-adjustment.yaml",
			}
			err := validate(alertMethod)

			if len(test.expectedErrors) == 0 {
				testutils.AssertNoError(t, test.spec, err)
			} else {
				testutils.AssertContainsErrors(t, test.spec, err, len(test.expectedErrors), test.expectedErrors...)
			}
		})
	}
}
