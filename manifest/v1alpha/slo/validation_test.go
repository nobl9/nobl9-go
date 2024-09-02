package slo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for SLO '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/slo.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	slo := validSLO()
	slo.APIVersion = "v0.1"
	slo.Kind = manifest.KindProject
	slo.ManifestSource = "/home/me/slo.yaml"
	err := validate(slo)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, slo, err, 2,
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
	slo := validSLO()
	slo.Metadata = Metadata{
		Name:        strings.Repeat("MY SLO", 20),
		DisplayName: strings.Repeat("my-slo", 20),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	slo.ManifestSource = "/home/me/slo.yaml"
	err := validate(slo)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, slo, err, 5,
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
	for name, test := range v1alphatest.GetLabelsTestCases[SLO](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validSLO()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[SLO](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validSLO()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Spec_Description(t *testing.T) {
	slo := validSLO()
	slo.Spec.Description = strings.Repeat("a", 2000)
	err := validate(slo)
	testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
		Prop: "spec.description",
		Code: validationV1Alpha.ErrorCodeStringDescription,
	})
}

func TestValidate_Spec_BudgetingMethod(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = BudgetingMethodOccurrences.String()
		testutils.AssertNoError(t, slo, validate(slo))
		slo.Spec.BudgetingMethod = BudgetingMethodTimeslices.String()
		slo.Spec.Objectives[0].TimeSliceTarget = ptr(0.1)
		testutils.AssertNoError(t, slo, validate(slo))
	})
	t.Run("empty method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.budgetingMethod",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.budgetingMethod",
			Message: "'invalid' is not a valid budgeting method",
		})
	})
}

func TestValidate_Spec_Service(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "my-service"
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "MY SERVICE"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.service",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
}

func TestValidate_Spec_AlertPolicies(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy"}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy", "MY POLICY", "ok-policy"}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.alertPolicies[1]",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
}

func TestValidate_Spec_Attachments(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, attachments := range [][]Attachment{
			{},
			{{URL: "https://my-url.com"}},
			{{URL: "https://my-url.com"}, {URL: "http://insecure-url.pl", DisplayName: ptr("Dashboard")}},
		} {
			slo := validSLO()
			slo.Spec.Attachments = attachments
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("fails, too many attachments", func(t *testing.T) {
		slo := validSLO()
		var attachments []Attachment
		for i := 0; i < 21; i++ {
			attachments = append(attachments, Attachment{})
		}
		slo.Spec.Attachments = attachments
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.attachments",
			Code: rules.ErrorCodeSliceLength,
		})
	})
	t.Run("fails, invalid attachment", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Attachments = []Attachment{
			{URL: "https://this.com"},
			{URL: ".com"},
			{URL: "", DisplayName: ptr(strings.Repeat("l", 64))},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.attachments[1].url",
				Code: rules.ErrorCodeStringURL,
			},
			testutils.ExpectedError{
				Prop: "spec.attachments[2].displayName",
				Code: rules.ErrorCodeStringLength,
			},
			testutils.ExpectedError{
				Prop: "spec.attachments[2].url",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
}

func TestValidate_Spec_Composite(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, composite := range []*Composite{
			nil,
			{
				BudgetTarget:      ptr(0.001),
				BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
			},
			{
				BudgetTarget:      ptr(0.9999),
				BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
			},
		} {
			slo := validSLO()
			slo.Spec.Composite = composite
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Composite     *Composite
			ExpectedError testutils.ExpectedError
		}{
			"target required": {
				Composite: &Composite{
					BudgetTarget:      nil,
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.target",
					Code: rules.ErrorCodeRequired,
				},
			},
			"target too small": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.target",
					Code: rules.ErrorCodeGreaterThan,
				},
			},
			"target too large": {
				Composite: &Composite{
					BudgetTarget:      ptr(1.0),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.target",
					Code: rules.ErrorCodeLessThan,
				},
			},
			"burn rate value too small": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: -1, Operator: "gt"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.burnRateCondition.value",
					Code: rules.ErrorCodeGreaterThanOrEqualTo,
				},
			},
			"burn rate value too large": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1001, Operator: "gt"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.burnRateCondition.value",
					Code: rules.ErrorCodeLessThanOrEqualTo,
				},
			},
			"missing operator": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 10},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.burnRateCondition.op",
					Code: rules.ErrorCodeRequired,
				},
			},
			"invalid operator": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 10, Operator: "lte"},
				},
				ExpectedError: testutils.ExpectedError{
					Prop: "spec.composite.burnRateCondition.op",
					Code: rules.ErrorCodeEqualTo,
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Composite = test.Composite
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, test.ExpectedError)
			})
		}
	})
	t.Run("missing burnRateCondition for occurrences", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = BudgetingMethodOccurrences.String()
		slo.Spec.Composite = &Composite{
			BudgetTarget:      ptr(0.9),
			BurnRateCondition: nil,
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.composite.burnRateCondition",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("burnRateCondition forbidden for timeslices", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = BudgetingMethodTimeslices.String()
		slo.Spec.Objectives[0].TimeSliceTarget = ptr(0.9)
		slo.Spec.Composite = &Composite{
			BudgetTarget:      ptr(0.9),
			BurnRateCondition: &CompositeBurnRateCondition{Value: 10, Operator: "gt"},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.composite.burnRateCondition",
			Code: rules.ErrorCodeForbidden,
		})
	})
}

func TestValidate_Spec_AnomalyConfig(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, config := range []*AnomalyConfig{
			nil,
			{NoData: nil},
			{NoData: &AnomalyConfigNoData{AlertMethods: []AnomalyConfigAlertMethod{{
				Name: "my-name",
			}}}},
			{NoData: &AnomalyConfigNoData{AlertMethods: []AnomalyConfigAlertMethod{{
				Name:    "my-name",
				Project: "default",
			}}}},
		} {
			slo := validSLO()
			slo.Spec.AnomalyConfig = config
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Config              *AnomalyConfig
			ExpectedErrors      []testutils.ExpectedError
			ExpectedErrorsCount int
		}{
			"no alert methods": {
				Config: &AnomalyConfig{NoData: &AnomalyConfigNoData{
					AlertMethods: make([]AnomalyConfigAlertMethod, 0),
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.anomalyConfig.noData.alertMethods",
						Code: rules.ErrorCodeSliceMinLength,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid name and project": {
				Config: &AnomalyConfig{NoData: &AnomalyConfigNoData{AlertMethods: []AnomalyConfigAlertMethod{
					{
						Name:    "",
						Project: "this-project",
					},
					{
						Name:    "MY NAME",
						Project: "THIS PROJECT",
					},
					{
						Name: "MY NAME",
					},
				}}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.anomalyConfig.noData.alertMethods[0].name",
						Code: rules.ErrorCodeRequired,
					},
					{
						Prop: "spec.anomalyConfig.noData.alertMethods[1].name",
						Code: rules.ErrorCodeStringDNSLabel,
					},
					{
						Prop: "spec.anomalyConfig.noData.alertMethods[1].project",
						Code: rules.ErrorCodeStringDNSLabel,
					},
					{
						Prop: "spec.anomalyConfig.noData.alertMethods[2].name",
						Code: rules.ErrorCodeStringDNSLabel,
					},
				},
				ExpectedErrorsCount: 4,
			},
			"not unique alert methods": {
				Config: &AnomalyConfig{NoData: &AnomalyConfigNoData{AlertMethods: []AnomalyConfigAlertMethod{
					{
						Name:    "my-name",
						Project: "default",
					},
					{
						Name:    "my-name",
						Project: "", // Will be filled with default.
					},
				}}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.anomalyConfig.noData.alertMethods",
						Code: rules.ErrorCodeSliceUnique,
					},
				},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.AnomalyConfig = test.Config
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

// nolint: lll
func TestValidate_Spec_TimeWindows(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, tw := range [][]TimeWindow{
			{TimeWindow{
				Unit:      "Day",
				Count:     1,
				IsRolling: true,
			}},
			{TimeWindow{
				Unit:      "Month",
				Count:     1,
				IsRolling: false,
				Calendar: &Calendar{
					StartTime: "2022-01-21 12:30:00",
					TimeZone:  "America/New_York",
				},
			}},
		} {
			slo := validSLO()
			slo.Spec.TimeWindows = tw
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			TimeWindows         []TimeWindow
			ExpectedErrors      []testutils.ExpectedError
			ExpectedErrorsCount int
		}{
			"no time windows": {
				TimeWindows: []TimeWindow{},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows",
						Code: rules.ErrorCodeSliceLength,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"too many time windows": {
				TimeWindows: []TimeWindow{{}, {}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows",
						Code: rules.ErrorCodeSliceLength,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"missing unit and count": {
				TimeWindows: []TimeWindow{{
					Unit:      "",
					Count:     0,
					IsRolling: true,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0].unit",
						Code: rules.ErrorCodeRequired,
					},
					{
						Prop: "spec.timeWindows[0].count",
						Code: rules.ErrorCodeGreaterThan,
					},
				},
				ExpectedErrorsCount: 2,
			},
			"invalid unit": {
				TimeWindows: []TimeWindow{{
					Unit:  "not-valid",
					Count: 1,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0].unit",
						Code: rules.ErrorCodeOneOf,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid calendar - missing fields": {
				TimeWindows: []TimeWindow{{
					Unit:     "Day",
					Count:    1,
					Calendar: &Calendar{},
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0].calendar.startTime",
						Code: rules.ErrorCodeRequired,
					},
					{
						Prop: "spec.timeWindows[0].calendar.timeZone",
						Code: rules.ErrorCodeRequired,
					},
				},
				ExpectedErrorsCount: 2,
			},
			"invalid calendar - invalid fields": {
				TimeWindows: []TimeWindow{{
					Unit:  "Day",
					Count: 1,
					Calendar: &Calendar{
						StartTime: "asd",
						TimeZone:  "asd",
					},
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop:    "spec.timeWindows[0].calendar.startTime",
						Message: `error parsing date: parsing time "asd" as "2006-01-02 15:04:05": cannot parse "asd" as "2006"`,
					},
					{
						Prop:    "spec.timeWindows[0].calendar.timeZone",
						Message: "not a valid time zone: unknown time zone asd",
					},
				},
				ExpectedErrorsCount: 2,
			},
			"isRolling and calendar are both set": {
				TimeWindows: []TimeWindow{{
					Unit:      "Day",
					Count:     1,
					IsRolling: true,
					Calendar: &Calendar{
						StartTime: "2022-01-21 12:30:00",
						TimeZone:  "America/New_York",
					},
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop:    "spec.timeWindows[0]",
						Message: "if 'isRolling' property is true, 'calendar' property must be omitted",
					},
				},
				ExpectedErrorsCount: 1,
			},
			"isRolling and calendar are both not set": {
				TimeWindows: []TimeWindow{{
					Unit:      "Day",
					Count:     1,
					IsRolling: false,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop:    "spec.timeWindows[0]",
						Message: "if 'isRolling' property is false or not set, 'calendar' property must be provided",
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid rolling time window unit": {
				TimeWindows: []TimeWindow{{
					Unit:      "Year",
					Count:     1,
					IsRolling: true,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop:    "spec.timeWindows[0]",
						Message: "invalid time window unit for Rolling window type: must be one of [Minute, Hour, Day]",
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid calendar time window unit": {
				TimeWindows: []TimeWindow{{
					Unit:      "Second",
					Count:     1,
					IsRolling: false,
					Calendar: &Calendar{
						StartTime: "2022-01-21 12:30:00",
						TimeZone:  "America/New_York",
					},
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop:    "spec.timeWindows[0]",
						Message: "invalid time window unit for Calendar window type: must be one of [Day, Week, Month, Quarter, Year]",
					},
				},
				ExpectedErrorsCount: 1,
			},
			"rolling time window size is less than defined min": {
				TimeWindows: []TimeWindow{{
					Unit:      "Minute",
					Count:     4,
					IsRolling: true,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0]",
						Message: fmt.Sprintf(
							"rolling time window size must be greater than or equal to %s",
							minimumRollingTimeWindowSize),
					},
				},
				ExpectedErrorsCount: 1,
			},
			"rolling time window size is greater than defined max": {
				TimeWindows: []TimeWindow{{
					Unit:      "Day",
					Count:     32,
					IsRolling: true,
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0]",
						Message: fmt.Sprintf(
							"rolling time window size must be less than or equal to %s",
							maximumRollingTimeWindowSize),
					},
				},
				ExpectedErrorsCount: 1,
			},
			"calendar time window size is greater than defined max": {
				TimeWindows: []TimeWindow{{
					Unit:      "Year",
					Count:     2,
					IsRolling: false,
					Calendar: &Calendar{
						StartTime: "2022-01-21 12:30:00",
						TimeZone:  "America/New_York",
					},
				}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.timeWindows[0]",
						Message: fmt.Sprintf(
							"calendar time window size must be less than %s",
							maximumCalendarTimeWindowSize),
					},
				},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.TimeWindows = test.TimeWindows
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec_Indicator(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, ind := range []Indicator{
			{
				MetricSource: MetricSourceSpec{Name: "name-only"},
			},
			{
				MetricSource: MetricSourceSpec{
					Name:    "name",
					Project: "default",
					Kind:    manifest.KindAgent,
				},
			},
			{
				MetricSource: MetricSourceSpec{
					Name:    "name",
					Project: "default",
					Kind:    manifest.KindDirect,
				},
			},
		} {
			slo := validSLO()
			slo.Spec.Indicator = &ind
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Indicator           *Indicator
			ExpectedErrors      []testutils.ExpectedError
			ExpectedErrorsCount int
		}{
			"no indicator": {
				Indicator: nil,
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.indicator",
						Code: rules.ErrorCodeRequired,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"empty indicator": {
				Indicator: &Indicator{},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.indicator.metricSource.name",
						Code: rules.ErrorCodeRequired,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"empty metric source name": {
				Indicator: &Indicator{MetricSource: MetricSourceSpec{Name: "", Project: "default"}},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.indicator.metricSource.name",
						Code: rules.ErrorCodeRequired,
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid metric source": {
				Indicator: &Indicator{
					MetricSource: MetricSourceSpec{
						Name:    "MY NAME",
						Project: "MY PROJECT",
						Kind:    manifest.KindSLO,
					},
				},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.indicator.metricSource.name",
						Code: rules.ErrorCodeStringDNSLabel,
					},
					{
						Prop: "spec.indicator.metricSource.project",
						Code: rules.ErrorCodeStringDNSLabel,
					},
					{
						Prop: "spec.indicator.metricSource.kind",
						Code: rules.ErrorCodeOneOf,
					},
				},
				ExpectedErrorsCount: 3,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Indicator = test.Indicator
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec_Objectives(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, objectives := range map[string][]Objective{
			"valid raw metric": {{
				ObjectiveBase: ObjectiveBase{
					Name:        "name",
					Value:       ptr(9.2),
					DisplayName: strings.Repeat("l", 63),
				},
				BudgetTarget: ptr(0.9),
				RawMetric:    &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
				Operator:     ptr(v1alpha.GreaterThan.String()),
			}},
			"empty value for count metrics": {{
				ObjectiveBase: ObjectiveBase{
					Name:        "name",
					Value:       nil,
					DisplayName: strings.Repeat("l", 63),
				},
				BudgetTarget: ptr(0.9),
				CountMetrics: &CountMetricsSpec{
					Incremental: ptr(false),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					GoodMetric:  validMetricSpec(v1alpha.Prometheus),
				},
				Operator: ptr(v1alpha.GreaterThan.String()),
			}},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = objectives
				err := validate(slo)
				testutils.AssertNoError(t, slo, err)
			})
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Objectives          []Objective
			ExpectedErrors      []testutils.ExpectedError
			ExpectedErrorsCount int
		}{
			"not enough objectives": {
				Objectives: []Objective{},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives",
						Code: rules.ErrorCodeSliceMinLength,
					},
				},
				ExpectedErrorsCount: 2,
			},
			"objective base errors": {
				Objectives: []Objective{
					{
						ObjectiveBase: ObjectiveBase{
							DisplayName: strings.Repeat("l", 64),
							Value:       ptr(2.),
							Name:        "MY NAME",
						},
						BudgetTarget: ptr(2.0),
						RawMetric:    &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:     ptr(v1alpha.GreaterThan.String()),
					},
					{
						ObjectiveBase: ObjectiveBase{Name: ""},
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(1.)},
						BudgetTarget:  ptr(-1.0),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					},
				},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].displayName",
						Code: rules.ErrorCodeStringMaxLength,
					},
					{
						Prop: "spec.objectives[0].name",
						Code: rules.ErrorCodeStringDNSLabel,
					},
					{
						Prop: "spec.objectives[0].target",
						Code: rules.ErrorCodeLessThan,
					},
					{
						Prop: "spec.objectives[1].value",
						Code: rules.ErrorCodeRequired,
					},
					{
						Prop: "spec.objectives[1].target",
						Code: rules.ErrorCodeRequired,
					},
					{
						Prop: "spec.objectives[2].target",
						Code: rules.ErrorCodeGreaterThanOrEqualTo,
					},
				},
				ExpectedErrorsCount: 6,
			},
			"all objectives have unique values": {
				Objectives: []Objective{
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(10.)},
						BudgetTarget:  ptr(0.9),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(10.)},
						BudgetTarget:  ptr(0.8),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					},
				},
				ExpectedErrors: []testutils.ExpectedError{{
					Prop: "spec.objectives",
					Code: rules.ErrorCodeSliceUnique,
				}},
				ExpectedErrorsCount: 1,
			},
			"invalid operator": {
				Objectives: []Objective{{
					ObjectiveBase: ObjectiveBase{Value: ptr(10.)},
					BudgetTarget:  ptr(0.9),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr("invalid"),
				}},
				ExpectedErrors: []testutils.ExpectedError{{
					Prop: "spec.objectives[0].op",
					Code: rules.ErrorCodeOneOf,
				}},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = test.Objectives
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec_Objectives_Primary(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for name, objectives := range map[string][]Objective{
			"no primary objective": {
				{
					ObjectiveBase: ObjectiveBase{Value: ptr(0.1)},
					BudgetTarget:  ptr(0.9),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr(v1alpha.GreaterThan.String()),
					Primary:       ptr(false),
				},
				{
					ObjectiveBase: ObjectiveBase{Value: ptr(0.2)},
					BudgetTarget:  ptr(0.8),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr(v1alpha.GreaterThan.String()),
				},
			},
			"exactly one primary objective": {
				{
					ObjectiveBase: ObjectiveBase{Value: ptr(0.1)},
					BudgetTarget:  ptr(0.9),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr(v1alpha.GreaterThan.String()),
					Primary:       ptr(true),
				},
				{
					ObjectiveBase: ObjectiveBase{Value: ptr(0.2)},
					BudgetTarget:  ptr(0.8),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr(v1alpha.GreaterThan.String()),
					Primary:       ptr(false),
				},
				{
					ObjectiveBase: ObjectiveBase{Value: ptr(0.3)},
					BudgetTarget:  ptr(0.7),
					RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					Operator:      ptr(v1alpha.GreaterThan.String()),
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = objectives
				err := validate(slo)
				testutils.AssertNoError(t, slo, err)
			})
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Objectives          []Objective
			ExpectedErrors      []testutils.ExpectedError
			ExpectedErrorsCount int
		}{
			"more than one primary objective": {
				Objectives: []Objective{
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(0.1)},
						BudgetTarget:  ptr(0.9),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
						Primary:       ptr(true),
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(0.2)},
						BudgetTarget:  ptr(0.8),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
						Primary:       ptr(true),
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(0.3)},
						BudgetTarget:  ptr(0.7),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					},
				},
				ExpectedErrors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives",
						Code: rules.ErrorCodeForbidden,
					},
				},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = test.Objectives
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec_Objectives_RawMetric(t *testing.T) {
	for name, test := range map[string]struct {
		Code       string
		InputValue float64
	}{
		"timeSliceTarget too low": {
			Code:       rules.ErrorCodeGreaterThan,
			InputValue: 0.0,
		},
		"timeSliceTarget too high": {
			Code:       rules.ErrorCodeLessThanOrEqualTo,
			InputValue: 1.1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			slo := validSLO()
			slo.Spec.BudgetingMethod = BudgetingMethodTimeslices.String()
			slo.Spec.Objectives[0].TimeSliceTarget = ptr(test.InputValue)
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: "spec.objectives[0].timeSliceTarget",
				Code: test.Code,
			})
		})
	}
}

func TestValidate_Spec(t *testing.T) {
	t.Run("exactly one metric type - both provided", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].RawMetric = &RawMetricSpec{
			MetricQuery: validMetricSpec(v1alpha.Prometheus),
		}
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			TotalMetric: validMetricSpec(v1alpha.Prometheus),
			GoodMetric:  validMetricSpec(v1alpha.Prometheus),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec",
				Code: errCodeExactlyOneMetricType,
			})
	})
	t.Run("exactly one metric type - both missing", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].RawMetric = nil
		slo.Spec.Objectives[0].CountMetrics = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec",
			Code: errCodeExactlyOneMetricType,
		})
	})
	t.Run("required time slice target for budgeting method", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.BudgetingMethod = BudgetingMethodTimeslices.String()
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].timeSliceTarget",
			Code: joinErrorCodes(errCodeTimeSliceTarget, rules.ErrorCodeRequired),
		})
	})
	t.Run("invalid time slice target for budgeting method", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.BudgetingMethod = BudgetingMethodOccurrences.String()
		slo.Spec.Objectives[0].TimeSliceTarget = ptr(0.1)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].timeSliceTarget",
			Code: joinErrorCodes(errCodeTimeSliceTarget, rules.ErrorCodeForbidden),
		})
	})
	t.Run("missing operator for raw metric", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].Operator = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].op",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_RawMetrics(t *testing.T) {
	t.Run("no metric spec provided", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Prometheus = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "must have exactly one metric spec type, none were provided",
		})
	})
	t.Run("exactly one metric spec type", func(t *testing.T) {
		for name, metrics := range map[string][]*MetricSpec{
			"single objective": {
				{
					Prometheus: validMetricSpec(v1alpha.Prometheus).Prometheus,
					Lightstep:  validMetricSpec(v1alpha.Lightstep).Lightstep,
				},
			},
			"two objectives": {
				validMetricSpec(v1alpha.Prometheus),
				validMetricSpec(v1alpha.Lightstep),
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = nil
				for i, m := range metrics {
					slo.Spec.Objectives = append(slo.Spec.Objectives, Objective{
						ObjectiveBase: ObjectiveBase{Value: ptr(10. + float64(i)), Name: strconv.Itoa(i)},
						BudgetTarget:  ptr(0.9),
						RawMetric:     &RawMetricSpec{MetricQuery: m},
						Operator:      ptr(v1alpha.GreaterThan.String()),
					})
				}
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: errCodeExactlyOneMetricSpecType,
				})
			})
		}
	})
	t.Run("query required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query",
			Code: rules.ErrorCodeRequired,
		})
	})
}

func TestValidate_Spec_CountMetrics(t *testing.T) {
	t.Run("no metric spec provided", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].CountMetrics.TotalMetric.Prometheus = nil
		slo.Spec.Objectives[0].CountMetrics.GoodMetric.Prometheus = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "must have exactly one metric spec type, none were provided",
		})
	})
	t.Run("bad over total enabled", func(t *testing.T) {
		for _, typ := range badOverTotalEnabledSources {
			slo := validSLO()
			slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
				Incremental: ptr(true),
				TotalMetric: validMetricSpec(typ),
				BadMetric:   validMetricSpec(typ),
			}
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("bad provided with good", func(t *testing.T) {
		slo := validCountMetricSLO(v1alpha.AzureMonitor)
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			TotalMetric: validMetricSpec(v1alpha.AzureMonitor),
			GoodMetric:  validMetricSpec(v1alpha.AzureMonitor),
			BadMetric:   validMetricSpec(v1alpha.AzureMonitor),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics",
			Code: errCodeEitherBadOrGoodCountMetric,
		})
	})
	t.Run("exactly one metric spec type", func(t *testing.T) {
		for name, metrics := range map[string][]*CountMetricsSpec{
			"single objective - total": {
				{
					Incremental: ptr(true),
					TotalMetric: &MetricSpec{
						Prometheus: validMetricSpec(v1alpha.Prometheus).Prometheus,
						Dynatrace:  validMetricSpec(v1alpha.Dynatrace).Dynatrace,
					},
					GoodMetric: validMetricSpec(v1alpha.Prometheus),
				},
			},
			"single objective - good": {
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					GoodMetric: &MetricSpec{
						Prometheus: validMetricSpec(v1alpha.Prometheus).Prometheus,
						Dynatrace:  validMetricSpec(v1alpha.Dynatrace).Dynatrace,
					},
				},
			},
			"single objective - bad": {
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.CloudWatch),
					BadMetric: &MetricSpec{
						CloudWatch:   validMetricSpec(v1alpha.CloudWatch).CloudWatch,
						AzureMonitor: validMetricSpec(v1alpha.AzureMonitor).AzureMonitor,
					},
				},
			},
			"single objective - good/total": {
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					GoodMetric:  validMetricSpec(v1alpha.Dynatrace),
				},
			},
			"single objective - bad/total": {
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					BadMetric:   validMetricSpec(v1alpha.CloudWatch),
				},
			},
			"two objectives - mix": {
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					GoodMetric:  validMetricSpec(v1alpha.Prometheus),
				},
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Prometheus),
					BadMetric:   validMetricSpec(v1alpha.CloudWatch),
				},
			},
			"two objectives - mix bad/total with single query good/total metric": {
				{
					Incremental:     ptr(true),
					GoodTotalMetric: validSingleQueryMetricSpec(v1alpha.Splunk),
				},
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Splunk),
					BadMetric:   validMetricSpec(v1alpha.CloudWatch),
				},
			},
			"two objectives - mix good/total with single query good/total metric": {
				{
					Incremental:     ptr(true),
					GoodTotalMetric: validSingleQueryMetricSpec(v1alpha.Splunk),
				},
				{
					Incremental: ptr(true),
					TotalMetric: validMetricSpec(v1alpha.Splunk),
					GoodMetric:  validMetricSpec(v1alpha.CloudWatch),
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = nil
				for i, m := range metrics {
					slo.Spec.Objectives = append(slo.Spec.Objectives, Objective{
						ObjectiveBase: ObjectiveBase{Value: ptr(10. + float64(i)), Name: strconv.Itoa(i)},
						BudgetTarget:  ptr(0.9),
						CountMetrics:  m,
					})
				}
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec",
					Code: errCodeExactlyOneMetricSpecType,
				})
			})
		}
	})
	t.Run("bad over total disabled", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			TotalMetric: validMetricSpec(v1alpha.Prometheus),
			BadMetric:   validMetricSpec(v1alpha.Prometheus),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].countMetrics.bad",
			Code: joinErrorCodes(errCodeBadOverTotalDisabled, rules.ErrorCodeOneOf),
		})
	})
}

func validRawMetricSLO(metricType v1alpha.DataSourceType) SLO {
	s := validSLO()
	s.Spec.Objectives[0].CountMetrics = nil
	s.Spec.Objectives[0].RawMetric = &RawMetricSpec{MetricQuery: validMetricSpec(metricType)}
	s.Spec.Objectives[0].TimeSliceTarget = nil
	s.Spec.Objectives[0].Operator = ptr(v1alpha.GreaterThan.String())
	return s
}

func validCountMetricSLO(metricType v1alpha.DataSourceType) SLO {
	s := validSLO()
	s.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
		Incremental: ptr(false),
		TotalMetric: validMetricSpec(metricType),
		GoodMetric:  validMetricSpec(metricType),
	}
	return s
}

// nolint:unparam
func validSingleQueryGoodOverTotalCountMetricSLO(metricType v1alpha.DataSourceType) SLO {
	s := validSLO()
	if !slices.Contains(singleQueryGoodOverTotalEnabledSources, metricType) {
		panic("metric type not supported")
	}
	s.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
		Incremental:     ptr(false),
		GoodTotalMetric: validSingleQueryMetricSpec(metricType),
	}
	return s
}

func validSLO() SLO {
	return New(
		Metadata{
			Name:        "my-slo",
			DisplayName: "My SLO",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		Spec{
			Description:   "Example slo",
			AlertPolicies: []string{"my-policy-name"},
			Attachments: []Attachment{
				{
					DisplayName: ptr("Grafana Dashboard"),
					URL:         "https://loki.my-org.dev/grafana/d/dnd48",
				},
			},
			BudgetingMethod: BudgetingMethodOccurrences.String(),
			Service:         "prometheus",
			Indicator: &Indicator{
				MetricSource: MetricSourceSpec{
					Project: "default",
					Name:    "prometheus",
					Kind:    manifest.KindAgent,
				},
			},
			Objectives: []Objective{
				{
					ObjectiveBase: ObjectiveBase{
						DisplayName: "Good",
						Value:       ptr(120.),
						Name:        "good",
					},
					BudgetTarget: ptr(0.9),
					CountMetrics: &CountMetricsSpec{
						Incremental: ptr(false),
						TotalMetric: validMetricSpec(v1alpha.Prometheus),
						GoodMetric:  validMetricSpec(v1alpha.Prometheus),
					},
					Operator: ptr(v1alpha.LessThan.String()),
				},
			},
			TimeWindows: []TimeWindow{
				{
					Unit:      "Day",
					Count:     1,
					IsRolling: true,
				},
			},
		},
	)
}

func validCompositeObjective() Objective {
	return Objective{
		ObjectiveBase: ObjectiveBase{
			DisplayName: "Composite",
			Value:       ptr(120.),
			Name:        "composite-1",
		},
		BudgetTarget: ptr(0.9),
		Composite: &CompositeSpec{
			MaxDelay: "10m",
			Components: Components{
				Objectives: []CompositeObjective{
					{
						Project:     "project-alpha",
						SLO:         "my-slo-alpha",
						Objective:   "good",
						Weight:      1.0,
						WhenDelayed: WhenDelayedCountAsGood,
					},
					{
						Project:     "project-beta",
						SLO:         "my-slo-beta",
						Objective:   "average",
						Weight:      2.0,
						WhenDelayed: WhenDelayedCountAsBad,
					},
				},
			},
		},
	}
}

func validCompositeSLO() SLO {
	return New(
		Metadata{
			Name:        "my-composite-slo",
			DisplayName: "My Composite SLO",
			Project:     "composite-project",
			Labels: v1alpha.Labels{
				"team":          []string{"blue", "red"},
				"business-unit": []string{"management"},
			},
		},
		Spec{
			Description:   "Example composite slo",
			AlertPolicies: []string{"my-policy-name"},
			Attachments: []Attachment{
				{
					DisplayName: ptr("Master report"),
					URL:         "https://example.com",
				},
			},
			BudgetingMethod: BudgetingMethodOccurrences.String(),
			Service:         "prometheus",
			Objectives: []Objective{
				validCompositeObjective(),
			},
			TimeWindows: []TimeWindow{
				{
					Unit:      "Day",
					Count:     1,
					IsRolling: true,
				},
			},
		},
	)
}

// Ensure that validateExactlyOneMetricSpecType function handles all possible data source types.
func TestValidateExactlyOneMetricSpecType(t *testing.T) {
	for _, s1 := range v1alpha.DataSourceTypeValues() {
		for _, s2 := range v1alpha.DataSourceTypeValues() {
			err := validateExactlyOneMetricSpecType(validMetricSpec(s1), validMetricSpec(s2))
			if s1 == s2 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}
	}
}

func validMetricSpec(typ v1alpha.DataSourceType) *MetricSpec {
	ms := validMetricSpecs[typ]
	var clone MetricSpec
	data, _ := json.Marshal(ms)
	_ = json.Unmarshal(data, &clone)
	return &clone
}

var validMetricSpecs = map[v1alpha.DataSourceType]MetricSpec{
	v1alpha.Prometheus: {Prometheus: &PrometheusMetric{
		PromQL: ptr(`sum(rate(prometheus_http_req`),
	}},
	v1alpha.Datadog: {Datadog: &DatadogMetric{
		Query: ptr(`avg:trace.http.request.duration{*}`),
	}},
	v1alpha.NewRelic: {NewRelic: &NewRelicMetric{
		NRQL: ptr(`SELECT average(duration*1000) FROM Transaction WHERE app Name='production' TIMESERIES`),
	}},
	v1alpha.AppDynamics: {AppDynamics: &AppDynamicsMetric{
		ApplicationName: ptr("my-app"),
		MetricPath:      ptr(`End User Experience|App|Slow Requests`),
	}},
	v1alpha.Splunk: {Splunk: &SplunkMetric{
		Query: ptr(`
search index=svc-events source=udp:5072 sourcetype=syslog status<400 |
bucket _time span=1m |
stats avg(response_time) as n9value by _time |
rename _time as n9time |
fields n9time n9value`),
	}},
	v1alpha.Lightstep: {Lightstep: &LightstepMetric{
		TypeOfData: ptr(LightstepMetricDataType),
		UQL:        ptr(`metric cpu.utilization | rate | group_by [], mean`),
	}},
	v1alpha.SplunkObservability: {SplunkObservability: &SplunkObservabilityMetric{
		Program: ptr(`data('demo.trans.count', filter=filter('demo_datacenter', 'Tokyo'), rollup='rate').mean().publish()`),
	}},
	v1alpha.Dynatrace: {Dynatrace: &DynatraceMetric{
		MetricSelector: ptr(`
builtin:synthetic.http.duration.geo
:filter(and(
  in("dt.entity.http_check",entitySelector("type(http_check),entityName(~"API Sample~")")),
  in("dt.entity.synthetic_location",entitySelector("type(synthetic_location),entityName(~"N. California~")"))))
:splitBy("dt.entity.http_check","dt.entity.synthetic_location")
:avg:auto:sort(value(avg,descending))
:limit(20)`),
	}},
	v1alpha.ThousandEyes: {ThousandEyes: &ThousandEyesMetric{
		TestID:   ptr(int64(2024796)),
		TestType: ptr("net-latency"),
	}},
	v1alpha.Graphite: {Graphite: &GraphiteMetric{
		MetricPath: ptr("stats.response.200"),
	}},
	v1alpha.BigQuery: {BigQuery: &BigQueryMetric{
		ProjectID: "svc-256112",
		Location:  "EU",
		Query: `
SELECT http_code AS n9value, created AS n9date
FROM 'test-256112.metrics.http_response'
WHERE http_code = 200 AND created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)`,
	}},
	v1alpha.Elasticsearch: {Elasticsearch: &ElasticsearchMetric{
		Index: ptr("apm-7.13.3-transaction"),
		Query: ptr(`
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "service.name": "service"
          }
        },
        {
          "match": {
            "transaction.result": "HTTP 2xx"
          }
        }
      ],
      "filter": [
        {
          "range": {
            "@timestamp": {
              "gte": "{{.BeginTime}}",
              "lte": "{{.EndTime}}"
            }
          }
        }
      ]
    }
  },
  "size": 0,
  "aggs": {
    "resolution": {
      "date_histogram": {
        "field": "@timestamp",
        "fixed_interval": "{{.Resolution}}",
        "min_doc_count": 0,
        "extended_bounds": {
          "min": "{{.BeginTime}}",
          "max": "{{.EndTime}}"
        }
      },
      "aggs": {
        "n9-val": {
          "avg": {
            "field": "transaction.duration.us"
          }
        }
      }
    }
  }
}
`),
	}},
	v1alpha.OpenTSDB: {OpenTSDB: &OpenTSDBMetric{
		Query: ptr(`m=avg:{{.N9RESOLUTION}}-avg:main_kafka_prometheus_go_memstats_alloc_bytes`),
	}},
	v1alpha.GrafanaLoki: {GrafanaLoki: &GrafanaLokiMetric{
		Logql: ptr(`
sum(sum_over_time({topic="error-budgets-out", consumergroup="alerts", cluster="main"} |=
"kafka_consumergroup_lag" |
logfmt |
kafka_consumergroup_lag!="" |
line_format "{{.kafka_consumergroup_lag}}" |
unwrap kafka_consumergroup_lag [1m]))`),
	}},
	v1alpha.CloudWatch: {CloudWatch: &CloudWatchMetric{
		Region:     ptr("eu-central-1"),
		Namespace:  ptr("AWS/Prometheus"),
		MetricName: ptr("CPUUtilization"),
		Stat:       ptr("Average"),
		Dimensions: []CloudWatchMetricDimension{
			{
				Name:  ptr("DBInstanceIdentifier"),
				Value: ptr("my-db-instance"),
			},
		},
		AccountID: ptr("123456789012"),
	}},
	v1alpha.Pingdom: {Pingdom: &PingdomMetric{
		CheckID:   ptr("8745322"),
		CheckType: ptr("uptime"),
		Status:    ptr("up"),
	}},
	v1alpha.AmazonPrometheus: {AmazonPrometheus: &AmazonPrometheusMetric{
		PromQL: ptr("sum(rate(prometheus_http_requests_total[1h]))"),
	}},
	v1alpha.Redshift: {Redshift: &RedshiftMetric{
		Region:       ptr("eu-central-1"),
		ClusterID:    ptr("my-redshift-cluster"),
		DatabaseName: ptr("my-database"),
		Query: ptr("SELECT value as n9value, timestamp as n9date FROM sinusoid" +
			" WHERE timestamp BETWEEN :n9date_from AND :n9date_to"),
	}},
	v1alpha.SumoLogic: {SumoLogic: &SumoLogicMetric{
		Type:         ptr("metrics"),
		Query:        ptr("kube_node_status_condition | min"),
		Quantization: ptr("1m"),
		Rollup:       ptr("Min"),
	}},
	v1alpha.Instana: {Instana: &InstanaMetric{
		MetricType: instanaMetricTypeInfrastructure,
		Infrastructure: &InstanaInfrastructureMetricType{
			MetricID:              "availableReplicas",
			PluginID:              "kubernetesDeployment",
			MetricRetrievalMethod: "query",
			Query:                 ptr("entity.kubernetes.namespace:kube-system AND entity.kubernetes.deployment.name:aws-load-balancer-controller"), //nolint:lll
		},
	}},
	v1alpha.InfluxDB: {InfluxDB: &InfluxDBMetric{
		Query: ptr(`
from(bucket: "integrations")
  |> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))
  |> aggregateWindow(every: 15s, fn: mean, createEmpty: false)
  |> filter(fn: (r) => r["_measurement"] == "internal_write")
  |> filter(fn: (r) => r["_field"] == "write_time_ns")
`),
	}},
	v1alpha.GCM: {GCM: &GCMMetric{
		Query: `
fetch consumed_api
  | metric 'serviceruntime.googleapis.com/api/request_count'
  | filter
      (resource.service == 'monitoring.googleapis.com')
      && (metric.response_code == '200')
  | align rate(1m)
  | every 1m
  | group_by [resource.service],
      [value_request_count_aggregate: aggregate(value.request_count)]
`,
		ProjectID: "svc-256112",
	}},
	v1alpha.AzureMonitor: {AzureMonitor: &AzureMonitorMetric{
		DataType:    AzureMonitorDataTypeMetrics,
		ResourceID:  "/subscriptions/9c26f90e/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
		MetricName:  "HttpResponseTime",
		Aggregation: "Avg",
	}},
	v1alpha.Generic: {Generic: &GenericMetric{
		Query: ptr("anything is valid"),
	}},
	v1alpha.Honeycomb: {Honeycomb: &HoneycombMetric{
		Calculation: "SUM",
		Attribute:   "http.status_code",
	}},
	v1alpha.LogicMonitor: {LogicMonitor: &LogicMonitorMetric{
		QueryType:                  "device_metrics",
		DeviceDataSourceInstanceID: 1029,
		GraphID:                    11354,
		Line:                       "MAXRTT",
	}},
	v1alpha.AzurePrometheus: {AzurePrometheus: &AzurePrometheusMetric{
		PromQL: "sum(rate(prometheus_http_requests_total[1h]))",
	}},
}

func validSingleQueryMetricSpec(typ v1alpha.DataSourceType) *MetricSpec {
	ms := validSingleQueryMetricSpecs[typ]
	var clone MetricSpec
	data, _ := json.Marshal(ms)
	_ = json.Unmarshal(data, &clone)
	return &clone
}

var validSingleQueryMetricSpecs = map[v1alpha.DataSourceType]MetricSpec{
	v1alpha.Splunk: {Splunk: &SplunkMetric{
		Query: ptr(`
    | mstats avg("spl.intr.resource_usage.IOWait.data.avg_cpu_pct") as n9good WHERE index="_metrics" span=15s
    | join type=left _time [
    | mstats avg("spl.intr.resource_usage.IOWait.data.max_cpus_pct") as n9total WHERE index="_metrics" span=15s
    ]
    | rename _time as n9time
    | fields n9time n9good n9total`),
	}},
}

func ptr[T any](v T) *T { return &v }

func joinErrorCodes(codes ...string) string {
	return strings.Join(codes, govy.ErrorCodeSeparator)
}
