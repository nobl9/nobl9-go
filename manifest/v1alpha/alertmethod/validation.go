package alertmethod

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	expectedNumberOfAlertMethodTypes = 1
	maxDescriptionLength             = 1050
	maxWebhookHeaders                = 10
	maxEmailRecipients               = 10
)

var (
	headerNameRegex     = regexp.MustCompile(`^([a-zA-Z0-9]+[_-]?)+$`)
	templateFieldsRegex = regexp.MustCompile(`\$([a-z_]+(\[])?)`)
)

func validate(a AlertMethod) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, a, manifest.KindAlertMethod)
}

var validator = govy.New[AlertMethod](
	validationV1Alpha.FieldRuleAPIVersion(func(a AlertMethod) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a AlertMethod) manifest.Kind { return a.Kind }, manifest.KindAlertMethod),
	govy.For(func(a AlertMethod) Metadata { return a.Metadata }).
		Include(metadataValidation),
	govy.For(func(a AlertMethod) Spec { return a.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = govy.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(m Metadata) v1alpha.MetadataAnnotations {
		return m.Annotations
	}),
)

var specValidation = govy.New[Spec](
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(rules.StringLength(0, maxDescriptionLength)),
	govy.For(govy.GetSelf[Spec]()).
		Rules(govy.NewRule(func(s Spec) error {
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
	govy.ForPointer(func(s Spec) *WebhookAlertMethod { return s.Webhook }).
		WithName("webhook").
		Include(webhookValidation),
	govy.ForPointer(func(s Spec) *PagerDutyAlertMethod { return s.PagerDuty }).
		WithName("pagerduty").
		Include(pagerDutyValidation),
	govy.ForPointer(func(s Spec) *SlackAlertMethod { return s.Slack }).
		WithName("slack").
		Include(slackValidation),
	govy.ForPointer(func(s Spec) *DiscordAlertMethod { return s.Discord }).
		WithName("discord").
		Include(discordValidation),
	govy.ForPointer(func(s Spec) *OpsgenieAlertMethod { return s.Opsgenie }).
		WithName("opsgenie").
		Include(opsgenieValidation),
	govy.ForPointer(func(s Spec) *ServiceNowAlertMethod { return s.ServiceNow }).
		WithName("servicenow").
		Include(serviceNowValidation),
	govy.ForPointer(func(s Spec) *JiraAlertMethod { return s.Jira }).
		WithName("jira").
		Include(jiraValidation),
	govy.ForPointer(func(s Spec) *TeamsAlertMethod { return s.Teams }).
		WithName("msteams").
		Include(teamsValidation),
	govy.ForPointer(func(s Spec) *EmailAlertMethod { return s.Email }).
		WithName("email").
		Include(emailValidation),
)

var webhookValidation = govy.New[WebhookAlertMethod](
	govy.For(govy.GetSelf[WebhookAlertMethod]()).
		Rules(
			govy.NewRule(func(w WebhookAlertMethod) error {
				if w.Template != nil && len(w.TemplateFields) > 0 {
					return errors.New("must not contain both 'template' and 'templateFields'")
				}
				if w.Template == nil && len(w.TemplateFields) == 0 {
					return errors.New("must contain either 'template' or 'templateFields'")
				}
				return nil
			})),
	govy.For(func(w WebhookAlertMethod) string { return w.URL }).
		WithName("url").
		HideValue().
		Include(optionalUrlValidation()),
	govy.ForPointer(func(w WebhookAlertMethod) *string { return w.Template }).
		WithName("template").
		Rules(govy.NewRule(func(v string) error {
			fields := extractTemplateFields(v)
			return validateTemplateFields(fields)
		})),
	govy.For(func(w WebhookAlertMethod) []string { return w.TemplateFields }).
		WithName("templateFields").
		OmitEmpty().
		Rules(govy.NewRule(validateTemplateFields)),
	govy.ForSlice(func(w WebhookAlertMethod) []WebhookHeader { return w.Headers }).
		WithName("headers").
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceMaxLength[[]WebhookHeader](maxWebhookHeaders)).
		IncludeForEach(webhookHeaderValidation),
)

var pagerDutyValidation = govy.New[PagerDutyAlertMethod](
	govy.For(func(p PagerDutyAlertMethod) string { return p.IntegrationKey }).
		WithName("integrationKey").
		HideValue().
		When(
			func(p PagerDutyAlertMethod) bool { return !isHiddenValue(p.IntegrationKey) },
			govy.WhenDescriptionf("is empty or equal to '%s'", v1alpha.HiddenValue),
		).
		Rules(rules.StringMaxLength(32)),
	govy.ForPointer(func(s PagerDutyAlertMethod) *SendResolution { return s.SendResolution }).
		WithName("sendResolution").
		Include(sendResolutionValidation),
)

var sendResolutionValidation = govy.New[SendResolution](
	govy.ForPointer(func(s SendResolution) *string { return s.Message }).
		WithName("message").
		OmitEmpty().
		Rules(rules.StringMaxLength(128)),
)

const validSlackURLPrefix = "https://hooks.slack.com/services/"

var slackValidation = govy.New[SlackAlertMethod](
	govy.For(func(s SlackAlertMethod) string { return s.URL }).
		WithName("url").
		HideValue().
		Include(optionalUrlWithPrefixValidation(validSlackURLPrefix)),
)

var discordValidation = govy.New[DiscordAlertMethod](
	govy.For(func(s DiscordAlertMethod) string { return s.URL }).
		WithName("url").
		HideValue().
		Cascade(govy.CascadeModeStop).
		Rules(
			govy.NewRule(func(v string) error {
				if strings.HasSuffix(strings.ToLower(v), "/slack") || strings.HasSuffix(strings.ToLower(v), "/github") {
					return errors.New("must not end with /slack or /github")
				}
				return nil
			})).
		Include(optionalUrlValidation()),
)

const (
	validOpsgenieURL   = "https://api.opsgenie.com"
	validOpsgenieEuURL = "https://api.eu.opsgenie.com"
)

var opsgenieValidation = govy.New[OpsgenieAlertMethod](
	govy.For(func(o OpsgenieAlertMethod) string { return o.URL }).
		WithName("url").
		Include(optionalUrlWithPrefixValidation(validOpsgenieURL, validOpsgenieEuURL)),
	govy.For(func(o OpsgenieAlertMethod) string { return o.Auth }).
		WithName("auth").
		HideValue().
		When(
			func(o OpsgenieAlertMethod) bool { return !isHiddenValue(o.Auth) },
			govy.WhenDescriptionf("is empty or equal to '%s'", v1alpha.HiddenValue),
		).
		Rules(
			govy.NewRule(func(v string) error {
				if !strings.HasPrefix(v, "Basic") &&
					!strings.HasPrefix(v, "GenieKey") {
					return errors.New("invalid auth format, should start with either GenieKey or Basic")
				}
				return nil
			})),
)

var serviceNowValidation = govy.New[ServiceNowAlertMethod](
	govy.For(func(s ServiceNowAlertMethod) string { return s.Username }).
		WithName("username").
		Required(),
	govy.For(func(s ServiceNowAlertMethod) string { return s.InstanceName }).
		WithName("instanceName").
		Required(),
	govy.ForPointer(func(s ServiceNowAlertMethod) *SendResolution { return s.SendResolution }).
		WithName("sendResolution").
		Include(sendResolutionValidation),
)

var jiraValidation = govy.New[JiraAlertMethod](
	govy.Transform(func(j JiraAlertMethod) string { return j.URL }, url.Parse).
		WithName("url").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.URL()).
		Rules(
			govy.NewRule(func(u *url.URL) error {
				if u.Scheme != "https" {
					return errors.New("requires https scheme")
				}
				return nil
			}),
		),
	govy.For(func(s JiraAlertMethod) string { return s.Username }).
		WithName("username").
		Required(),
	govy.For(func(s JiraAlertMethod) string { return s.ProjectKey }).
		WithName("projectKey").
		Required(),
)

var teamsValidation = govy.New[TeamsAlertMethod](
	govy.Transform(func(t TeamsAlertMethod) string { return t.URL }, url.Parse).
		WithName("url").
		HideValue().
		Cascade(govy.CascadeModeStop).
		Rules(rules.URL()).
		Rules(
			govy.NewRule(func(u *url.URL) error {
				if u.Scheme != "https" {
					return errors.New("requires https scheme")
				}
				return nil
			}),
		),
).When(
	func(v TeamsAlertMethod) bool { return !isHiddenValue(v.URL) },
	govy.WhenDescriptionf("is empty or equal to '%s'", v1alpha.HiddenValue),
)

var emailValidation = govy.New[EmailAlertMethod](
	govy.For(govy.GetSelf[EmailAlertMethod]()).
		Rules(
			govy.NewRule(func(e EmailAlertMethod) error {
				if len(e.To) == 0 && len(e.Cc) == 0 && len(e.Bcc) == 0 {
					return errors.New("must contain at least one recipient")
				}
				return nil
			})),
	govy.For(func(s EmailAlertMethod) []string { return s.To }).
		WithName("to").
		Rules(rules.SliceMaxLength[[]string](maxEmailRecipients)),
	govy.For(func(s EmailAlertMethod) []string { return s.Cc }).
		WithName("cc").
		Rules(rules.SliceMaxLength[[]string](maxEmailRecipients)),
	govy.For(func(s EmailAlertMethod) []string { return s.Bcc }).
		WithName("bcc").
		Rules(rules.SliceMaxLength[[]string](maxEmailRecipients)),
)

func optionalUrlWithPrefixValidation(prefixes ...string) govy.Validator[string] {
	return govy.New[string](
		govy.For(govy.GetSelf[string]()).
			When(
				func(v string) bool { return !isHiddenValue(v) },
				govy.WhenDescriptionf("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(rules.StringURL(), rules.StringStartsWith(prefixes...)),
	)
}

func optionalUrlValidation() govy.Validator[string] {
	return govy.New[string](
		govy.For(govy.GetSelf[string]()).
			When(
				func(v string) bool { return !isHiddenValue(v) },
				govy.WhenDescriptionf("is empty or equal to '%s'", v1alpha.HiddenValue),
			).
			Rules(rules.StringURL()),
	)
}

var webhookHeaderValidation = govy.New[WebhookHeader](
	govy.For(func(h WebhookHeader) string { return h.Name }).
		WithName("name").
		Required().
		Rules(
			rules.StringNotEmpty(),
			rules.StringMatchRegexp(headerNameRegex).
				WithDetails("must be a valid header name")),
	govy.For(govy.GetSelf[WebhookHeader]()).
		Include(
			webhookHeaderValueValidation,
			webhookHeaderSecretValueValidation),
)

var webhookHeaderValueValidation = govy.New[WebhookHeader](
	govy.For(func(h WebhookHeader) string { return h.Value }).
		WithName("value").
		Required().
		Rules(rules.StringNotEmpty()),
).When(
	func(h WebhookHeader) bool { return !h.IsSecret },
	govy.WhenDescription("isSecret is false"),
)

var webhookHeaderSecretValueValidation = govy.New[WebhookHeader](
	govy.For(func(h WebhookHeader) string { return h.Value }).
		WithName("value").
		HideValue().
		Required().
		Rules(rules.StringNotEmpty()),
).When(
	func(h WebhookHeader) bool { return h.IsSecret },
	govy.WhenDescription("isSecret is true"),
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
