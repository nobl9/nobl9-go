package slo

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_metadata_error.txt
var expectedMetadataError string

func TestValidate_Metadata(t *testing.T) {
	slo := validSLO()
	slo.Metadata = Metadata{
		Name:        strings.Repeat("MY SLO", 20),
		DisplayName: strings.Repeat("my-slo", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
		Labels: v1alpha.Labels{
			"L O L": []string{"dip", "dip"},
		},
	}
	slo.Spec.Description = strings.Repeat("l", 2000)
	slo.ManifestSource = "/home/me/slo.yaml"
	err := validate(slo)
	require.Error(t, err)
	assert.EqualError(t, err, expectedMetadataError)
	fmt.Println(err)
}

func TestValidate_Spec_BudgetingMethod(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		for _, method := range []string{
			BudgetingMethodOccurrences.String(),
			BudgetingMethodTimeslices.String(),
		} {
			slo.Spec.BudgetingMethod = method
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("empty method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = ""
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec.budgetingMethod",
			Codes: []string{validation.ErrorCodeRequired},
		})
	})
	t.Run("invalid method", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.BudgetingMethod = "invalid"
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:     "spec.budgetingMethod",
			Messages: []string{"'invalid' is not a valid budgeting method"},
		})
	})
}

func TestValidate_Spec_Service(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "my-service"
		err := validate(slo)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Service = "MY SERVICE"
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec.service",
			Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
		})
	})
}

func TestValidate_Spec_AlertPolicies(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy"}
		err := validate(slo)
		assert.NoError(t, err)
	})
	t.Run("fails", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.AlertPolicies = []string{"my-policy", "MY POLICY", "ok-policy"}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec.alertPolicies[1]",
			Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
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
			assert.NoError(t, err)
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
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec.attachments",
			Codes: []string{validation.ErrorCodeSliceLength},
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
		assertContainsErrors(t, err, 3,
			expectedError{
				Prop:  "spec.attachments[1].url",
				Codes: []string{validation.ErrorCodeStringURL},
			},
			expectedError{
				Prop:  "spec.attachments[2].displayName",
				Codes: []string{validation.ErrorCodeStringLength},
			},
			expectedError{
				Prop:  "spec.attachments[2].url",
				Codes: []string{validation.ErrorCodeRequired},
			},
		)
	})
}

func TestValidate_Spec_Composite(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, composite := range []*Composite{
			nil,
			{
				BudgetTarget: ptr(0.001),
			},
			{
				BudgetTarget:      ptr(0.9999),
				BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
			},
		} {
			slo := validSLO()
			slo.Spec.Composite = composite
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Composite     *Composite
			ExpectedError expectedError
		}{
			"target required": {
				Composite: &Composite{
					BudgetTarget:      nil,
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.target",
					Codes: []string{validation.ErrorCodeRequired},
				},
			},
			"target too small": {
				Composite: &Composite{BudgetTarget: ptr(0.)},
				ExpectedError: expectedError{
					Prop:  "spec.composite.target",
					Codes: []string{validation.ErrorCodeGreaterThan},
				},
			},
			"target too large": {
				Composite: &Composite{BudgetTarget: ptr(1.0)},
				ExpectedError: expectedError{
					Prop:  "spec.composite.target",
					Codes: []string{validation.ErrorCodeLessThan},
				},
			},
			"burn rate value too small": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: -1, Operator: "gt"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.value",
					Codes: []string{validation.ErrorCodeGreaterThanOrEqualTo},
				},
			},
			"burn rate value too large": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1001, Operator: "gt"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.value",
					Codes: []string{validation.ErrorCodeLessThanOrEqualTo},
				},
			},
			"missing operator": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 10},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.op",
					Codes: []string{validation.ErrorCodeRequired},
				},
			},
			"invalid operator": {
				Composite: &Composite{
					BudgetTarget:      ptr(0.9),
					BurnRateCondition: &CompositeBurnRateCondition{Value: 10, Operator: "lte"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.op",
					Codes: []string{validation.ErrorCodeOneOf},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Composite = test.Composite
				err := validate(slo)
				assertContainsErrors(t, err, 1, test.ExpectedError)
			})
		}
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
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Config              *AnomalyConfig
			ExpectedErrors      []expectedError
			ExpectedErrorsCount int
		}{
			"no alert methods": {
				Config: &AnomalyConfig{NoData: &AnomalyConfigNoData{
					AlertMethods: make([]AnomalyConfigAlertMethod, 0),
				}},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods",
						Codes: []string{validation.ErrorCodeSliceMinLength},
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
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods[0].name",
						Codes: []string{validation.ErrorCodeRequired},
					},
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods[1].name",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
					},
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods[1].project",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
					},
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods[2].name",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
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
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.anomalyConfig.noData.alertMethods",
						Codes: []string{validation.ErrorCodeSliceUnique},
					},
				},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.AnomalyConfig = test.Config
				err := validate(slo)
				assertContainsErrors(t, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
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
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			TimeWindows         []TimeWindow
			ExpectedErrors      []expectedError
			ExpectedErrorsCount int
		}{
			"no time windows": {
				TimeWindows: []TimeWindow{},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.timeWindows",
						Codes: []string{validation.ErrorCodeSliceLength},
					},
				},
				ExpectedErrorsCount: 1,
			},
			"too many time windows": {
				TimeWindows: []TimeWindow{{}, {}},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.timeWindows",
						Codes: []string{validation.ErrorCodeSliceLength},
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
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.timeWindows[0].unit",
						Codes: []string{validation.ErrorCodeRequired},
					},
					{
						Prop:  "spec.timeWindows[0].count",
						Codes: []string{validation.ErrorCodeGreaterThan},
					},
				},
				ExpectedErrorsCount: 2,
			},
			"invalid unit": {
				TimeWindows: []TimeWindow{{
					Unit:  "dayz",
					Count: 1,
				}},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.timeWindows[0].unit",
						Codes: []string{validation.ErrorCodeOneOf},
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
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.timeWindows[0].calendar.startTime",
						Codes: []string{validation.ErrorCodeRequired},
					},
					{
						Prop:  "spec.timeWindows[0].calendar.timeZone",
						Codes: []string{validation.ErrorCodeRequired},
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
				ExpectedErrors: []expectedError{
					{
						Prop:     "spec.timeWindows[0].calendar.startTime",
						Messages: []string{`error parsing date: parsing time "asd" as "2006-01-02 15:04:05": cannot parse "asd" as "2006"`},
					},
					{
						Prop:     "spec.timeWindows[0].calendar.timeZone",
						Messages: []string{"not a valid time zone: unknown time zone asd"},
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
				ExpectedErrors: []expectedError{
					{
						Prop:     "spec.timeWindows[0]",
						Messages: []string{"if 'isRolling' property is true, 'calendar' property must be omitted"},
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
				ExpectedErrors: []expectedError{
					{
						Prop:     "spec.timeWindows[0]",
						Messages: []string{"if 'isRolling' property is false or not set, 'calendar' property must be provided"},
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
				ExpectedErrors: []expectedError{
					{
						Prop:     "spec.timeWindows[0]",
						Messages: []string{"invalid time window unit for Rolling window type: must be one of [Minute, Hour, Day]"},
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
				ExpectedErrors: []expectedError{
					{
						Prop:     "spec.timeWindows[0]",
						Messages: []string{"invalid time window unit for Calendar window type: must be one of [Day, Week, Month, Quarter, Year]"},
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
				ExpectedErrors: []expectedError{
					{
						Prop: "spec.timeWindows[0]",
						Messages: []string{fmt.Sprintf(
							"rolling time window size must be greater than or equal to %s",
							minimumRollingTimeWindowSize)},
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
				ExpectedErrors: []expectedError{
					{
						Prop: "spec.timeWindows[0]",
						Messages: []string{fmt.Sprintf(
							"rolling time window size must be less than or equal to %s",
							maximumRollingTimeWindowSize)},
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
				ExpectedErrors: []expectedError{
					{
						Prop: "spec.timeWindows[0]",
						Messages: []string{fmt.Sprintf(
							"calendar time window size must be less than %s",
							maximumCalendarTimeWindowSize)},
					},
				},
				ExpectedErrorsCount: 1,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.TimeWindows = test.TimeWindows
				err := validate(slo)
				assertContainsErrors(t, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
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
			slo.Spec.Indicator = ind
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Indicator           Indicator
			ExpectedErrors      []expectedError
			ExpectedErrorsCount int
		}{
			"empty indicator": {
				Indicator: Indicator{},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.indicator",
						Codes: []string{validation.ErrorCodeRequired},
					},
				},
				ExpectedErrorsCount: 1,
			},
			"empty metric source name": {
				Indicator: Indicator{MetricSource: MetricSourceSpec{Name: "", Project: "default"}},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.indicator.metricSource.name",
						Codes: []string{validation.ErrorCodeRequired},
					},
				},
				ExpectedErrorsCount: 1,
			},
			"invalid metric source": {
				Indicator: Indicator{
					MetricSource: MetricSourceSpec{
						Name:    "MY NAME",
						Project: "MY PROJECT",
						Kind:    manifest.KindSLO,
					},
				},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.indicator.metricSource.name",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
					},
					{
						Prop:  "spec.indicator.metricSource.project",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
					},
					{
						Prop:  "spec.indicator.metricSource.kind",
						Codes: []string{validation.ErrorCodeOneOf},
					},
				},
				ExpectedErrorsCount: 3,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Indicator = test.Indicator
				err := validate(slo)
				assertContainsErrors(t, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec_Objectives(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		for _, objectives := range [][]Objective{
			{{
				ObjectiveBase: ObjectiveBase{
					Name:        "name",
					Value:       ptr(9.2),
					DisplayName: strings.Repeat("l", 63),
				},
				BudgetTarget: ptr(0.9),
				RawMetric:    &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
			}},
		} {
			slo := validSLO()
			slo.Spec.Objectives = objectives
			err := validate(slo)
			assert.NoError(t, err)
		}
	})
	t.Run("fails", func(t *testing.T) {
		for name, test := range map[string]struct {
			Objectives          []Objective
			ExpectedErrors      []expectedError
			ExpectedErrorsCount int
		}{
			"not enough objectives": {
				Objectives: []Objective{},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.objectives",
						Codes: []string{validation.ErrorCodeSliceMinLength},
					},
				},
				ExpectedErrorsCount: 2,
			},
			"objective base errors": {
				Objectives: []Objective{
					{
						ObjectiveBase: ObjectiveBase{
							DisplayName: strings.Repeat("l", 64),
							Value:       ptr(0.),
							Name:        "MY NAME",
						},
						BudgetTarget: ptr(2.0),
						RawMetric:    &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					},
					{
						ObjectiveBase: ObjectiveBase{Name: ""},
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(0.)},
						BudgetTarget:  ptr(-1.0),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					},
				},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.objectives[0].displayName",
						Codes: []string{validation.ErrorCodeStringMaxLength},
					},
					{
						Prop:  "spec.objectives[0].name",
						Codes: []string{validation.ErrorCodeStringIsDNSSubdomain},
					},
					{
						Prop:  "spec.objectives[0].target",
						Codes: []string{validation.ErrorCodeLessThan},
					},
					{
						Prop:  "spec.objectives[1].value",
						Codes: []string{validation.ErrorCodeRequired},
					},
					{
						Prop:  "spec.objectives[1].target",
						Codes: []string{validation.ErrorCodeRequired},
					},
					{
						Prop:  "spec.objectives[2].target",
						Codes: []string{validation.ErrorCodeGreaterThanOrEqualTo},
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
					},
					{
						ObjectiveBase: ObjectiveBase{Value: ptr(10.)},
						BudgetTarget:  ptr(0.8),
						RawMetric:     &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
					},
				},
				ExpectedErrors: []expectedError{
					{
						Prop:  "spec.objectives",
						Codes: []string{validation.ErrorCodeSliceUnique},
					},
				},
				ExpectedErrorsCount: 2,
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = test.Objectives
				err := validate(slo)
				assertContainsErrors(t, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
			})
		}
	})
}

func TestValidate_Spec(t *testing.T) {
	metricSpec := &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("sum")}}
	t.Run("exactly one metric type - both provided", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].RawMetric = &RawMetricSpec{MetricQuery: metricSpec}
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			TotalMetric: metricSpec,
			GoodMetric:  metricSpec,
		}
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec",
			Codes: []string{errCodeExactlyOneMetricType},
		})
	})
	t.Run("exactly one metric type - both missing", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].RawMetric = nil
		slo.Spec.Objectives[0].CountMetrics = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec",
			Codes: []string{errCodeExactlyOneMetricType},
		})
	})
}

func TestValidate_Spec_RawMetrics(t *testing.T) {
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
						ObjectiveBase: ObjectiveBase{Value: ptr(10.), Name: strconv.Itoa(i)},
						BudgetTarget:  ptr(0.9),
						RawMetric:     &RawMetricSpec{MetricQuery: m},
					})
				}
				err := validate(slo)
				assertContainsErrors(t, err, 1, expectedError{
					Prop:  "spec",
					Codes: []string{errCodeExactlyOneMetricSpecType},
				})
			})
		}
	})
}

func TestValidate_Spec_CountMetrics(t *testing.T) {
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
		} {
			t.Run(name, func(t *testing.T) {
				slo := validSLO()
				slo.Spec.Objectives = nil
				for i, m := range metrics {
					slo.Spec.Objectives = append(slo.Spec.Objectives, Objective{
						ObjectiveBase: ObjectiveBase{Value: ptr(10.), Name: strconv.Itoa(i)},
						BudgetTarget:  ptr(0.9),
						CountMetrics:  m,
					})
				}
				err := validate(slo)
				assertContainsErrors(t, err, 1, expectedError{
					Prop:  "spec",
					Codes: []string{errCodeExactlyOneMetricSpecType},
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
		assertContainsErrors(t, err, 1, expectedError{
			Prop:  "spec.objectives[0].countMetrics.bad",
			Codes: []string{errCodeBadOverTotalDisabled},
		})
	})
}

type expectedError struct {
	Prop     string
	Codes    []string
	Messages []string
}

func assertContainsErrors(t *testing.T, err error, expectedErrorsCount int, expectedErrors ...expectedError) {
	t.Helper()
	// Convert to ObjectError.
	require.Error(t, err)
	var objErr *v1alpha.ObjectError
	require.ErrorAs(t, err, &objErr)
	// Count errors.
	actualErrorsCount := 0
	for _, actual := range objErr.Errors {
		var propErr *validation.PropertyError
		require.ErrorAs(t, actual, &propErr)
		actualErrorsCount += len(propErr.Errors)
	}
	require.Equalf(t,
		expectedErrorsCount,
		actualErrorsCount,
		"%T contains a different number of errors than expected", err)
	// Find and match expected errors.
	for _, expected := range expectedErrors {
		found := false
	searchErrors:
		for _, actual := range objErr.Errors {
			var propErr *validation.PropertyError
			require.ErrorAs(t, actual, &propErr)
			if propErr.PropertyName != expected.Prop {
				continue
			}
			for _, actualRuleErr := range propErr.Errors {
				for _, expectedMessage := range expected.Messages {
					if expectedMessage == actualRuleErr.Message {
						found = true
						break searchErrors
					}
				}
				for _, expectedCode := range expected.Codes {
					if expectedCode == actualRuleErr.Code {
						found = true
						break searchErrors
					}
				}
			}
		}
		require.Truef(t, found, "expected '%v' error was not found", expected)
	}
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
			Indicator: Indicator{
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

func validMetricSpec(typ v1alpha.DataSourceType) *MetricSpec {
	ms := validMetricSpecs[typ]
	return &ms
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
		Query:     "SELECT http_code AS n9value, created AS n9date FROM `bdwtest-256112.metrics.http_response` WHERE http_code = 200 AND created BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)",
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
            "service.name": "weloveourpets_xyz"
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
		Query:        ptr("SELECT value as n9value, timestamp as n9date FROM sinusoid WHERE timestamp BETWEEN :n9date_from AND :n9date_to"),
	}},
	v1alpha.SumoLogic: {SumoLogic: &SumoLogicMetric{
		Type:         ptr("metrics"),
		Query:        ptr("kube_node_status_condition | min"),
		Quantization: ptr("1m"),
		Rollup:       ptr("Min"),
	}},
	v1alpha.Instana: {Instana: &InstanaMetric{
		MetricType: instanaMetricTypeApplication,
		Application: &InstanaApplicationMetricType{
			MetricID:    "latency",
			Aggregation: "p99",
			GroupBy: InstanaApplicationMetricGroupBy{
				Tag:       "endpoint.name",
				TagEntity: "DESTINATION",
			},
			APIQuery: `
{
  "type": "EXPRESSION",
  "logicalOperator": "AND",
  "elements": [
    {
      "type": "TAG_FILTER",
      "name": "service.name",
      "operator": "EQUALS",
      "entity": "DESTINATION",
      "value": "master"
    },
    {
      "type": "TAG_FILTER",
      "name": "call.type",
      "operator": "EQUALS",
      "entity": "NOT_APPLICABLE",
      "value": "HTTP"
    }
  ]
}
`,
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
		ResourceID:  "/subscriptions/9c26f90e/resourceGroups/azure-monitor-test-sources/providers/Microsoft.Web/sites/app",
		MetricName:  "HttpResponseTime",
		Aggregation: "Avg",
	}},
}

func ptr[T any](v T) *T { return &v }
