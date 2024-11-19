package alertmethod

// TemplateVariable can be used with MessageRenderer to override how variables are rendered.
type TemplateVariable = string

const (
	TplVarProjectName                     TemplateVariable = "project_name"
	TplVarServiceName                     TemplateVariable = "service_name"
	TplVarServiceLabelsText               TemplateVariable = "service_labels_text"
	TplVarOrganization                    TemplateVariable = "organization"
	TplVarAlertPolicyName                 TemplateVariable = "alert_policy_name"
	TplVarAlertPolicyLabelsText           TemplateVariable = "alert_policy_labels_text"
	TplVarAlertPolicyDescription          TemplateVariable = "alert_policy_description"
	TplVarAlertPolicyConditionsArray      TemplateVariable = "alert_policy_conditions[]"
	TplVarAlertPolicyConditionsText       TemplateVariable = "alert_policy_conditions_text"
	TplVarSeverity                        TemplateVariable = "severity"
	TplVarSloName                         TemplateVariable = "slo_name"
	TplVarSloLabelsText                   TemplateVariable = "slo_labels_text"
	TplVarSloDetailsLink                  TemplateVariable = "slo_details_link"
	TplVarObjectiveName                   TemplateVariable = "objective_name"
	TplVarTimestamp                       TemplateVariable = "timestamp"
	TplVarIsoTimestamp                    TemplateVariable = "iso_timestamp"
	TplVarAlertID                         TemplateVariable = "alert_id"
	TplVarBackwardCompatibleObjectiveName TemplateVariable = "experience_name"
)

var notificationTemplateAllowedFields = map[TemplateVariable]struct{}{
	TplVarProjectName:                     {},
	TplVarServiceName:                     {},
	TplVarServiceLabelsText:               {},
	TplVarOrganization:                    {},
	TplVarAlertPolicyName:                 {},
	TplVarAlertPolicyLabelsText:           {},
	TplVarAlertPolicyDescription:          {},
	TplVarAlertPolicyConditionsArray:      {},
	TplVarAlertPolicyConditionsText:       {},
	TplVarSeverity:                        {},
	TplVarSloName:                         {},
	TplVarSloLabelsText:                   {},
	TplVarSloDetailsLink:                  {},
	TplVarObjectiveName:                   {},
	TplVarTimestamp:                       {},
	TplVarIsoTimestamp:                    {},
	TplVarAlertID:                         {},
	TplVarBackwardCompatibleObjectiveName: {},
}
