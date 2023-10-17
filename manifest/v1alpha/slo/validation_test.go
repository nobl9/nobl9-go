package slo

import (
	_ "embed"
	"fmt"
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
				BudgetTarget: 0.001,
			},
			{
				BudgetTarget:      0.9999,
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
			"target too small": {
				Composite: &Composite{BudgetTarget: 0},
				ExpectedError: expectedError{
					Prop:  "spec.composite.target",
					Codes: []string{validation.ErrorCodeGreaterThan},
				},
			},
			"target too large": {
				Composite: &Composite{BudgetTarget: 1.0},
				ExpectedError: expectedError{
					Prop:  "spec.composite.target",
					Codes: []string{validation.ErrorCodeLessThan},
				},
			},
			"burn rate value too small": {
				Composite: &Composite{
					BudgetTarget:      0.9,
					BurnRateCondition: &CompositeBurnRateCondition{Value: -1, Operator: "gt"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.value",
					Codes: []string{validation.ErrorCodeGreaterThanOrEqualTo},
				},
			},
			"burn rate value too large": {
				Composite: &Composite{
					BudgetTarget:      0.9,
					BurnRateCondition: &CompositeBurnRateCondition{Value: 1001, Operator: "gt"},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.value",
					Codes: []string{validation.ErrorCodeLessThanOrEqualTo},
				},
			},
			"missing operator": {
				Composite: &Composite{
					BudgetTarget:      0.9,
					BurnRateCondition: &CompositeBurnRateCondition{Value: 10},
				},
				ExpectedError: expectedError{
					Prop:  "spec.composite.burnRateCondition.op",
					Codes: []string{validation.ErrorCodeRequired},
				},
			},
			"invalid operator": {
				Composite: &Composite{
					BudgetTarget:      0.9,
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
				Config: &AnomalyConfig{NoData: &AnomalyConfigNoData{}},
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
					Unit:  "",
					Count: 0,
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
						DisplayName: "",
						Value:       0,
						Name:        "",
						NameChanged: false,
					},
					BudgetTarget: ptr(0.9),
					CountMetrics: &CountMetricsSpec{
						Incremental: ptr(false),
						GoodMetric: &MetricSpec{
							Prometheus: &PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total{code=~"^2.*"}[1h]))`),
							},
						},
						TotalMetric: &MetricSpec{
							Prometheus: &PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total[1h]))`),
							},
						},
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

func ptr[T any](v T) *T { return &v }
