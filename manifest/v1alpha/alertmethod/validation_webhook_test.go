package alertmethod

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
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
			Template: getStringPointer("{\"slo\": \"$slo_name\"}"),
		},
		"passes with valid template fields": {
			TemplateFields: getAllowedTemplateFields(),
		},
		"passes with valid headers": {
			Template: &template,
			Headers: []WebhookHeader{
				{
					Name:     "Origin",
					Value:    "http://example.com",
					IsSecret: false,
				},
			},
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
					Code: validation.ErrorCodeStringURL,
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
					Message: "contains invalid template fields",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:      "http://example.com",
				Template: getStringPointer("{\"slo\": \"$sloname\"}"),
			},
		},
		"fails with unsupported template fields": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook.templateFields",
					Message: "contains invalid template fields",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:            "http://example.com",
				TemplateFields: []string{"$sloname"},
			},
		},
		"fails with both template and template fields defined": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.webhook",
					Message: "must not contain both template and templateFields",
				},
			},
			AlertMethod: WebhookAlertMethod{
				URL:            "http://example.com",
				Template:       &template,
				TemplateFields: []string{"$slo_name"},
			},
		},
		"fails with empty header name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].name",
					Code: validation.ErrorCodeRequired,
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
		"fails with empty header value": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.webhook.headers[0].value",
					Code: validation.ErrorCodeRequired,
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
