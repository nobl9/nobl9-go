package alertmethod

import (
	"fmt"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidate_Spec_WebhookAlertMethod(t *testing.T) {
	template := "i'm an alert"
	for name, spec := range map[string]WebhookAlertMethod{
		"passes with valid http url": {
			URL:      "http://example.com",
			Template: &template,
		},
		"passes with valid https url": {
			URL:      "https://example.com",
			Template: &template,
		},
		"passes with undefined url": {
			Template: &template,
		},
		"passes with empty url": {
			URL:      "",
			Template: &template,
		},
		"passes with hidden url": {
			URL:      v1alpha.HiddenValue,
			Template: &template,
		},
		"passes with valid template": {
			Template: ptr("{\"slo\": \"$slo_name\"}"),
		},
		"passes with valid template for no data anomaly": {
			Template: ptr("{\"no_data_alert_after\": \"no_data_alert_after\"}"),
		},
		"passes with valid template fields": {
			TemplateFields: getAllowedTemplateFields(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Webhook: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         WebhookAlertMethod
	}{
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.url",
					Code: rules.ErrorCodeStringURL,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "example.com",
				Template: &template,
			},
		},
		"fails with invalid template": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook.template",
					Message: "contains invalid template field: sloname",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: ptr("{\"slo\": \"$sloname\"}"),
			},
		},
		"fails with unsupported template fields": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook.templateFields",
					Message: "contains invalid template field: sloname",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:            "http://example.com",
				TemplateFields: []string{"sloname"},
			},
		},
		"fails with both template and template fields defined": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook",
					Message: "must not contain both 'template' and 'templateFields'",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:            "http://example.com",
				Template:       &template,
				TemplateFields: []string{"slo_name"},
			},
		},
		"fails with no template and template fields": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook",
					Message: "must contain either 'template' or 'templateFields'",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL: "http://example.com",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Webhook: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func TestValidate_Spec_WebhookHeaders(t *testing.T) {
	template := "i'm an alert"
	for name, spec := range map[string]WebhookAlertMethod{
		"passes with valid headers": {
			Template: &template,
			Headers: []WebhookHeader{{
				Name:     "Origin",
				Value:    "http://example.com",
				IsSecret: false,
			}},
		},
		"passes with proper valid headers length": {
			Template: &template,
			Headers:  generateHeaders(10),
		},
		"secret passes with hidden value": {
			Template: &template,
			Headers: []WebhookHeader{{
				Name:     "Origin",
				Value:    v1alpha.HiddenValue,
				IsSecret: true,
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Webhook: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         WebhookAlertMethod
	}{
		"fails with empty header name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].name",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: &template,
				Headers: []WebhookHeader{
					{
						Name:     "",
						Value:    "http://example.com",
						IsSecret: false,
					},
				},
			},
		},
		"fails with invalid header name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].name",
					Code: rules.ErrorCodeStringMatchRegexp,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: &template,
				Headers: []WebhookHeader{
					{
						Name:     " 42",
						Value:    "http://example.com",
						IsSecret: false,
					},
				},
			},
		},
		"fails with empty header value": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].value",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: &template,
				Headers: []WebhookHeader{
					{
						Name:     "Origin",
						IsSecret: false,
					},
				},
			},
		},
		"fails with empty secret header value": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].value",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: &template,
				Headers: []WebhookHeader{
					{
						Name:     "Origin",
						IsSecret: false,
					},
				},
			},
		},
		"fails with headers max length": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers",
					Code: rules.ErrorCodeSliceMaxLength,
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: &template,
				Headers:  generateHeaders(11),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Webhook: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func getAllowedTemplateFields() (result []string) {
	for field := range notificationTemplateAllowedFields {
		result = append(result, field)
	}
	return result
}

func generateHeaders(number int) []WebhookHeader {
	var headers []WebhookHeader
	for i := 0; i < number; i++ {
		headers = append(headers, WebhookHeader{
			Name:  fmt.Sprintf("%v", i),
			Value: fmt.Sprintf("value %v", i),
		})
	}
	return headers
}
