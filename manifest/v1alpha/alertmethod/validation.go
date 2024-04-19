package alertmethod

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	expectedNumberOfAlertMethodTypes = 1
	maxDescriptionLength             = 1050
	maxWebhookHeaders                = 10
	maxEmailRecipients               = 10
)

var headerNameRegex = regexp.MustCompile(`^([a-zA-Z0-9]+[_-]?)+$`)
var templateFieldsRegex = regexp.MustCompile(`\$([a-z_]+(\[])?)`)

var validator = validation.New[AlertMethod](
	validation.For(func(a AlertMethod) Metadata { return a.Metadata }).
		Include(metadataValidation),
	validation.For(func(a AlertMethod) Spec { return a.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringLength(0, maxDescriptionLength)),
	validation.For(validation.GetSelf[Spec]()).
		Rules(validation.NewSingleRule(func(s Spec) error {
			alertMethodCounter := 0
			if s.Webhook != nil {
				alertMethodCounter++
			}
			if s.PagerDuty != nil {
				alertMethodCounter++
			}
			if s.Slack != nil {
				alertMethodCounter++
			}
			if s.Discord != nil {
				alertMethodCounter++
			}
			if s.Opsgenie != nil {
				alertMethodCounter++
			}
			if s.ServiceNow != nil {
				alertMethodCounter++
			}
			if s.Jira != nil {
				alertMethodCounter++
			}
			if s.Teams != nil {
				alertMethodCounter++
			}
			if s.Email != nil {
				alertMethodCounter++
			}
			if alertMethodCounter != expectedNumberOfAlertMethodTypes {
				return errors.New("exactly one alert method configuration is required")
			}
			return nil
		})),
	validation.ForPointer(func(s Spec) *WebhookAlertMethod { return s.Webhook }).
		WithName("webhook").
		Include(webhookValidation),
	validation.ForPointer(func(s Spec) *PagerDutyAlertMethod { return s.PagerDuty }).
		WithName("pagerduty").
		Include(pagerDutyValidation),
	validation.ForPointer(func(s Spec) *SlackAlertMethod { return s.Slack }).
		WithName("slack").
		Include(slackValidation),
	validation.ForPointer(func(s Spec) *DiscordAlertMethod { return s.Discord }).
		WithName("discord").
		Include(discordValidation),
	validation.ForPointer(func(s Spec) *OpsgenieAlertMethod { return s.Opsgenie }).
		WithName("opsgenie").
		Include(opsgenieValidation),
	validation.ForPointer(func(s Spec) *ServiceNowAlertMethod { return s.ServiceNow }).
		WithName("servicenow").
		Include(serviceNowValidation),
	validation.ForPointer(func(s Spec) *JiraAlertMethod { return s.Jira }).
		WithName("jira").
		Include(jiraValidation),
	validation.ForPointer(func(s Spec) *TeamsAlertMethod { return s.Teams }).
		WithName("msteams").
		Include(teamsValidation),
	validation.ForPointer(func(s Spec) *EmailAlertMethod { return s.Email }).
		WithName("email").
		Include(emailValidation),
)

var webhookValidation = validation.New[WebhookAlertMethod](
	validation.For(validation.GetSelf[WebhookAlertMethod]()).
		Rules(
			validation.NewSingleRule(func(w WebhookAlertMethod) error {
				if w.Template != nil && len(w.TemplateFields) > 0 {
					return errors.New("must not contain both 'template' and 'templateFields'")
				}
				if w.Template == nil && len(w.TemplateFields) == 0 {
					return errors.New("must contain either 'template' or 'templateFields'")
				}
				return nil
			})),
	validation.For(func(w WebhookAlertMethod) string { return w.URL }).
		WithName("url").
		HideValue().
		Include(optionalUrlValidation()),
	validation.ForPointer(func(w WebhookAlertMethod) *string { return w.Template }).
		WithName("template").
		Rules(validation.NewSingleRule(func(v string) error {
			fields := extractTemplateFields(v)
			return validateTemplateFields(fields)
		})),
	validation.For(func(w WebhookAlertMethod) []string { return w.TemplateFields }).
		WithName("templateFields").
		OmitEmpty().
		Rules(validation.NewSingleRule(validateTemplateFields)),
	validation.ForSlice(func(w WebhookAlertMethod) []WebhookHeader { return w.Headers }).
		WithName("headers").
		Cascade(validation.CascadeModeStop).
		Rules(validation.SliceMaxLength[[]WebhookHeader](maxWebhookHeaders)).
		IncludeForEach(webhookHeaderValidation),
)

var pagerDutyValidation = validation.New[PagerDutyAlertMethod](
	validation.For(func(p PagerDutyAlertMethod) string { return p.IntegrationKey }).
		WithName("integrationKey").
		HideValue().
		When(
			func(p PagerDutyAlertMethod) bool { return !isHiddenValue(p.IntegrationKey) },
			validation.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
		).
		Rules(validation.StringMaxLength(32)),
)

var slackValidation = validation.New[SlackAlertMethod](
	validation.For(func(s SlackAlertMethod) string { return s.URL }).
		WithName("url").
		HideValue().
		Include(optionalUrlValidation()),
)

var discordValidation = validation.New[DiscordAlertMethod](
	validation.For(func(s DiscordAlertMethod) string { return s.URL }).
		WithName("url").
		HideValue().
		Cascade(validation.CascadeModeStop).
		Rules(
			validation.NewSingleRule(func(v string) error {
				if strings.HasSuffix(strings.ToLower(v), "/slack") || strings.HasSuffix(strings.ToLower(v), "/github") {
					return errors.New("must not end with /slack or /github")
				}
				return nil
			})).
		Include(optionalUrlValidation()),
)

var opsgenieValidation = validation.New[OpsgenieAlertMethod](
	validation.For(func(o OpsgenieAlertMethod) string { return o.URL }).
		WithName("url").
		Include(optionalUrlValidation()),
	validation.For(func(o OpsgenieAlertMethod) string { return o.Auth }).
		WithName("auth").
		HideValue().
		When(
			func(o OpsgenieAlertMethod) bool { return !isHiddenValue(o.Auth) },
			validation.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
		).
		Rules(
			validation.NewSingleRule(func(v string) error {
				if !strings.HasPrefix(v, "Basic") &&
					!strings.HasPrefix(v, "GenieKey") {
					return errors.New("invalid auth format, should start with either GenieKey or Basic")
				}
				return nil
			})),
)

var serviceNowValidation = validation.New[ServiceNowAlertMethod](
	validation.For(func(s ServiceNowAlertMethod) string { return s.Username }).
		WithName("username").
		Required(),
	validation.For(func(s ServiceNowAlertMethod) string { return s.InstanceName }).
		WithName("instanceName").
		Required(),
)

var jiraValidation = validation.New[JiraAlertMethod](
	validation.Transform(func(j JiraAlertMethod) string { return j.URL }, url.Parse).
		WithName("url").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.URL()).
		Rules(
			validation.NewSingleRule(func(u *url.URL) error {
				if u.Scheme != "https" {
					return errors.New("requires https scheme")
				}
				return nil
			}),
		),
	validation.For(func(s JiraAlertMethod) string { return s.Username }).
		WithName("username").
		Required(),
	validation.For(func(s JiraAlertMethod) string { return s.ProjectKey }).
		WithName("projectKey").
		Required(),
)

var teamsValidation = validation.New[TeamsAlertMethod](
	validation.Transform(func(t TeamsAlertMethod) string { return t.URL }, url.Parse).
		WithName("url").
		HideValue().
		Cascade(validation.CascadeModeStop).
		Rules(validation.URL()).
		Rules(
			validation.NewSingleRule(func(u *url.URL) error {
				if u.Scheme != "https" {
					return errors.New("requires https scheme")
				}
				return nil
			}),
		),
).When(
	func(v TeamsAlertMethod) bool { return !isHiddenValue(v.URL) },
	validation.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
)

var emailValidation = validation.New[EmailAlertMethod](
	validation.For(validation.GetSelf[EmailAlertMethod]()).
		Rules(
			validation.NewSingleRule(func(e EmailAlertMethod) error {
				if len(e.To) == 0 && len(e.Cc) == 0 && len(e.Bcc) == 0 {
					return errors.New("must contain at least one recipient")
				}
				return nil
			})),
	validation.For(func(s EmailAlertMethod) []string { return s.To }).
		WithName("to").
		Rules(validation.SliceMaxLength[[]string](maxEmailRecipients)),
	validation.For(func(s EmailAlertMethod) []string { return s.Cc }).
		WithName("cc").
		Rules(validation.SliceMaxLength[[]string](maxEmailRecipients)),
	validation.For(func(s EmailAlertMethod) []string { return s.Bcc }).
		WithName("bcc").
		Rules(validation.SliceMaxLength[[]string](maxEmailRecipients)),
)

func optionalUrlValidation() validation.Validator[string] {
	return validation.New[string](
		validation.For(validation.GetSelf[string]()).
			When(
				func(v string) bool { return !isHiddenValue(v) },
				validation.WhenDescription("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(validation.StringURL()),
	)
}

var webhookHeaderValidation = validation.New[WebhookHeader](
	validation.For(func(h WebhookHeader) string { return h.Name }).
		WithName("name").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMatchRegexp(headerNameRegex).
				WithDetails("must be a valid header name")),
	validation.For(validation.GetSelf[WebhookHeader]()).
		Include(
			webhookHeaderValueValidation,
			webhookHeaderSecretValueValidation),
)

var webhookHeaderValueValidation = validation.New[WebhookHeader](
	validation.For(func(h WebhookHeader) string { return h.Value }).
		WithName("value").
		Required().
		Rules(validation.StringNotEmpty()),
).When(
	func(h WebhookHeader) bool { return !h.IsSecret },
	validation.WhenDescription("isSecret is false"),
)

var webhookHeaderSecretValueValidation = validation.New[WebhookHeader](
	validation.For(func(h WebhookHeader) string { return h.Value }).
		WithName("value").
		HideValue().
		Required().
		Rules(validation.StringNotEmpty()),
).When(
	func(h WebhookHeader) bool { return h.IsSecret },
	validation.WhenDescription("isSecret is true"),
)

func extractTemplateFields(template string) []string {
	matches := templateFieldsRegex.FindAllStringSubmatch(template, -1)
	templateFields := make([]string, len(matches))
	for i, match := range matches {
		templateFields[i] = match[1]
	}
	return templateFields
}

func validateTemplateFields(templateFields []string) error {
	for _, field := range templateFields {
		if _, ok := notificationTemplateAllowedFields[field]; !ok {
			return errors.New("contains invalid template field: " + field)
		}
	}
	return nil
}

func isHiddenValue(s string) bool { return s == "" || s == v1alpha.HiddenValue }

func validate(a AlertMethod) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, a)
}
