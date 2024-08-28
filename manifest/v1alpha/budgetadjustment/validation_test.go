package budgetadjustment

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
	testutils.AssertContainsErrors(t, budgetAdjustment, err, 3,
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
				FirstEventStart: time.Now(),
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
				FirstEventStart: time.Now(),
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
					Code: rules.ErrorCodeDurationPrecision,
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
					Code: rules.ErrorCodeDurationPrecision,
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
					Code: rules.ErrorCodeRequired,
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
					Code: rules.ErrorCodeStringDNSLabel,
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
					Code: rules.ErrorCodeRequired,
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
					Code: rules.ErrorCodeStringDNSLabel,
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
			name: "duplicate slo",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeSliceUnique,
				},
			},
		},
		{
			name: "duplicate slo with multiple others",
			spec: Spec{
				FirstEventStart: time.Now(),
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
					Code: validation.ErrorCodeSliceUnique,
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
			FirstEventStart: time.Now(),
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
