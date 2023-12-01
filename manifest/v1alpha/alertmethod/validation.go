package alertmethod

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

const (
	// hiddenValue can be used as a value of a secret field and is ignored during saving
	hiddenValue                      = "[hidden]"
	expectedNumberOfAlertMethodTypes = 1
	maxDescriptionLength             = 1050
	maxWebhookHeaders                = 10
	maxEmailReceipients              = 10
)

var HeaderNameRegex = regexp.MustCompile(`^([a-zA-Z0-9]+[_-]?)+$`)

var alertMethodValidation = validation.New[AlertMethod](
	validation.For(func(a AlertMethod) Metadata { return a.Metadata }).
		Include(metadataValidation),
	validation.For(func(a AlertMethod) Spec { return a.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
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
					return errors.New("must not contain both template and templateFields")
				}
				if w.Template == nil && len(w.TemplateFields) == 0 {
					return errors.New("must contain either template or templateFields")
				}
				return nil
			})),
	validation.For(func(w WebhookAlertMethod) string { return w.URL }).
		WithName("url").
		Include(optionalUrlValidation()),
	validation.ForPointer(func(w WebhookAlertMethod) *string { return w.Template }).
		WithName("template").
		When(func(w WebhookAlertMethod) bool { return w.Template != nil }).
		Rules(webhookTemplateValidationRule()),
	validation.For(func(w WebhookAlertMethod) []string { return w.TemplateFields }).
		WithName("templateFields").
		When(func(w WebhookAlertMethod) bool { return w.Template == nil }).
		Rules(validation.SliceMinLength[[]string](1)).
		StopOnError().
		Rules(webhookTemplateFieldsValidationRule()),
	validation.ForEach(func(w WebhookAlertMethod) []WebhookHeader { return w.Headers }).
		WithName("headers").
		Rules(validation.SliceMaxLength[[]WebhookHeader](maxWebhookHeaders)).
		StopOnError().
		IncludeForEach(webhookHeaderValidation),
)

var pagerDutyValidation = validation.New[PagerDutyAlertMethod](
	validation.For(func(p PagerDutyAlertMethod) string { return p.IntegrationKey }).
		WithName("integrationKey").
		Rules(validation.StringMaxLength(32)).
		When(func(p PagerDutyAlertMethod) bool { return p.IntegrationKey != "" && p.IntegrationKey != hiddenValue }),
)

var slackValidation = validation.New[SlackAlertMethod](
	validation.For(func(s SlackAlertMethod) string { return s.URL }).
		WithName("url").
		Include(optionalUrlValidation()),
)

var discordValidation = validation.New[DiscordAlertMethod](
	validation.For(func(s DiscordAlertMethod) string { return s.URL }).
		WithName("url").
		Required().
		Rules(
			validation.NewSingleRule(func(v string) error {
				if strings.HasSuffix(strings.ToLower(v), "/slack") || strings.HasSuffix(strings.ToLower(v), "/github") {
					return errors.New("must not end with /slack or /github")
				}
				return nil
			})).
		StopOnError().
		Include(optionalUrlValidation()),
)

var opsgenieValidation = validation.New[OpsgenieAlertMethod](
	validation.For(func(s OpsgenieAlertMethod) string { return s.URL }).
		WithName("url").
		Include(optionalUrlValidation()),
	validation.For(func(s OpsgenieAlertMethod) string { return s.Auth }).
		WithName("auth").
		Include(opsgenieAuthValidation),
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
	validation.For(func(s JiraAlertMethod) string { return s.URL }).
		WithName("url").
		Required().
		Rules(
			validation.StringURL(),
		),
	validation.For(func(s JiraAlertMethod) string { return s.Username }).
		WithName("username").
		Required(),
	validation.For(func(s JiraAlertMethod) string { return s.ProjectKey }).
		WithName("projectKey").
		Required(),
)

var teamsValidation = validation.New[TeamsAlertMethod](
	validation.For(func(s TeamsAlertMethod) string { return s.URL }).
		WithName("url").
		Include(optionalUrlValidation(validation.RequireHttps)),
)

var emailValidation = validation.New[EmailAlertMethod](
	validation.For(validation.GetSelf[EmailAlertMethod]()).
		WithName("email").
		Rules(
			validation.NewSingleRule(func(e EmailAlertMethod) error {
				if len(e.To) == 0 && len(e.Cc) == 0 && len(e.Bcc) == 0 {
					return errors.New("must contain at least one recipient")
				}
				return nil
			})),
	validation.For(func(s EmailAlertMethod) []string { return s.To }).
		WithName("to").
		Rules(validation.SliceMaxLength[[]string](maxEmailReceipients)),
	validation.For(func(s EmailAlertMethod) []string { return s.Cc }).
		WithName("cc").
		Rules(validation.SliceMaxLength[[]string](maxEmailReceipients)),
	validation.For(func(s EmailAlertMethod) []string { return s.Bcc }).
		WithName("bcc").
		Rules(validation.SliceMaxLength[[]string](maxEmailReceipients)),
)

func optionalUrlValidation(options ...validation.StringURLOption) validation.Validator[string] {
	return validation.New[string](
		validation.For(validation.GetSelf[string]()).
			When(func(v string) bool { return v != "" && v != hiddenValue }).
			Rules(validation.StringURL(options...)),
	)
}

func webhookTemplateValidationRule() validation.SingleRule[string] {
	return validation.NewSingleRule(func(v string) error {
		templateFields := extractTemplateFields(v)
		if !hasValidTemplateFields(templateFields, notificationTemplateAllowedFields) {
			return errors.New("contains invalid template fields")
		}
		return nil
	})
}

func webhookTemplateFieldsValidationRule() validation.SingleRule[[]string] {
	return validation.NewSingleRule(func(v []string) error {
		if !hasValidTemplateFields(v, notificationTemplateAllowedFields) {
			return errors.New("contains invalid template fields")
		}
		return nil
	})
}

var webhookHeaderValidation = validation.New[WebhookHeader](
	validation.For(func(h WebhookHeader) string { return h.Name }).
		WithName("name").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMatchRegexp(HeaderNameRegex).
				WithDetails("must be a valid header name")),
	validation.For(func(h WebhookHeader) string { return h.Value }).
		WithName("value").
		Required().
		Rules(validation.StringNotEmpty()),
)

func extractTemplateFields(template string) []string {
	matches := regexp.MustCompile(`\$([a-z_]+(\[])?)`).FindAllStringSubmatch(template, -1)
	templateFields := make([]string, len(matches))
	for i, match := range matches {
		templateFields[i] = match[1]
	}
	return templateFields
}

func hasValidTemplateFields(templateFields []string, allowedFields map[string]struct{}) bool {
	for _, field := range templateFields {
		if _, ok := allowedFields[field]; !ok {
			return false
		}
	}
	return true
}

var opsgenieAuthValidation = validation.New[string](
	validation.For(validation.GetSelf[string]()).
		WithName("auth").
		Rules(
			validation.NewSingleRule(func(v string) error {
				if !strings.HasPrefix(v, "Basic") &&
					!strings.HasPrefix(v, "GenieKey") {
					return errors.New("invalid auth format")
				}
				return nil
			})).
		When(func(v string) bool { return v != "" && v != hiddenValue }))

func validate(a AlertMethod) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertMethodValidation, a)
}
