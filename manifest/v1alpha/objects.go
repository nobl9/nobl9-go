// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// APIVersion is a value of valid apiVersions
const APIVersion = "n9/v1alpha"

// APIObjects - all Objects available for this version of API
// Sorted in order of applying
type APIObjects struct {
	SLOs          SLOsSlice          `json:"slos,omitempty"`
	Services      ServicesSlice      `json:"services,omitempty"`
	Agents        AgentsSlice        `json:"agents,omitempty"`
	AlertPolicies AlertPoliciesSlice `json:"alertpolicies,omitempty"`
	AlertSilences AlertSilencesSlice `json:"alertsilences,omitempty"`
	Alerts        AlertsSlice        `json:"alerts,omitempty"`
	AlertMethods  AlertMethodsSlice  `json:"alertmethods,omitempty"`
	Directs       DirectsSlice       `json:"directs,omitempty"`
	DataExports   DataExportsSlice   `json:"dataexports,omitempty"`
	Projects      ProjectsSlice      `json:"projects,omitempty"`
	RoleBindings  RoleBindingsSlice  `json:"rolebindings,omitempty"`
	Annotations   AnnotationsSlice   `json:"annotations,omitempty"`
	UserGroups    UserGroupsSlice    `json:"usergroups,omitempty"`
}

func (o APIObjects) Clone() APIObjects {
	return APIObjects{
		SLOs:          o.SLOs.Clone(),
		Services:      o.Services.Clone(),
		Agents:        o.Agents.Clone(),
		AlertPolicies: o.AlertPolicies.Clone(),
		AlertSilences: o.AlertSilences.Clone(),
		Alerts:        o.Alerts.Clone(),
		AlertMethods:  o.AlertMethods.Clone(),
		Directs:       o.Directs.Clone(),
		DataExports:   o.DataExports.Clone(),
		Projects:      o.Projects.Clone(),
		RoleBindings:  o.RoleBindings.Clone(),
		Annotations:   o.Annotations.Clone(),
	}
}

func (o APIObjects) Len() int {
	return len(o.SLOs) +
		len(o.Services) +
		len(o.Agents) +
		len(o.AlertPolicies) +
		len(o.AlertSilences) +
		len(o.Alerts) +
		len(o.AlertMethods) +
		len(o.Directs) +
		len(o.DataExports) +
		len(o.Projects) +
		len(o.RoleBindings) +
		len(o.Annotations)
}

type ObjectSpec struct {
	ApiVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
}

func (o ObjectSpec) GetVersion() string {
	return o.ApiVersion
}

func (o ObjectSpec) GetKind() manifest.Kind {
	return o.Kind
}

// FilterEntry represents single metric label to be matched against value
type FilterEntry struct {
	Label string `json:"label" validate:"required,prometheusLabelName"`
	Value string `json:"value" validate:"required"`
}

type OrganizationInformation struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"displayName"`
}

// Applying multiple Agents at once can cause timeout for whole sloctl apply command.
// This is caused by long request to Okta to create client credentials app.
// The same case is applicable for delete command.
const allowedAgentsToModify = 1

// Parse takes care of all Object supported by n9/v1alpha apiVersion
func Parse(o manifest.ObjectGeneric, parsedObjects *APIObjects, onlyHeaders bool) (err error) {
	v := NewValidator()
	switch o.Kind {
	case manifest.KindSLO:
		var slo SLO
		slo, err = genericToSLO(o, v, onlyHeaders)
		parsedObjects.SLOs = append(parsedObjects.SLOs, slo)
	case manifest.KindService:
		var service Service
		service, err = genericToService(o, v, onlyHeaders)
		parsedObjects.Services = append(parsedObjects.Services, service)
	case manifest.KindAgent:
		var agent Agent
		if len(parsedObjects.Agents) >= allowedAgentsToModify {
			err = manifest.EnhanceError(o, errors.New("only one Agent can be defined in this configuration"))
		} else {
			agent, err = genericToAgent(o, v, onlyHeaders)
			parsedObjects.Agents = append(parsedObjects.Agents, agent)
		}
	case manifest.KindAlertPolicy:
		var alertPolicy AlertPolicy
		alertPolicy, err = genericToAlertPolicy(o, v, onlyHeaders)
		parsedObjects.AlertPolicies = append(parsedObjects.AlertPolicies, alertPolicy)
	case manifest.KindAlertSilence:
		var alertSilence AlertSilence
		alertSilence, err = genericToAlertSilence(o, v, onlyHeaders)
		parsedObjects.AlertSilences = append(parsedObjects.AlertSilences, alertSilence)
	case manifest.KindAlertMethod:
		var alertMethod AlertMethod
		alertMethod, err = genericToAlertMethod(o, v, onlyHeaders)
		parsedObjects.AlertMethods = append(parsedObjects.AlertMethods, alertMethod)
	case manifest.KindDirect:
		var direct Direct
		direct, err = genericToDirect(o, v, onlyHeaders)
		parsedObjects.Directs = append(parsedObjects.Directs, direct)
	case manifest.KindDataExport:
		var dataExport DataExport
		dataExport, err = genericToDataExport(o, v, onlyHeaders)
		parsedObjects.DataExports = append(parsedObjects.DataExports, dataExport)
	case manifest.KindProject:
		var project Project
		project, err = genericToProject(o, v, onlyHeaders)
		parsedObjects.Projects = append(parsedObjects.Projects, project)
	case manifest.KindRoleBinding:
		var roleBinding RoleBinding
		roleBinding, err = genericToRoleBinding(o, v)
		parsedObjects.RoleBindings = append(parsedObjects.RoleBindings, roleBinding)
	case manifest.KindAnnotation:
		var annotation Annotation
		annotation, err = genericToAnnotation(o, v)
		parsedObjects.Annotations = append(parsedObjects.Annotations, annotation)
	case manifest.KindUserGroup:
		var group UserGroup
		group, err = genericToUserGroup(o)
		parsedObjects.UserGroups = append(parsedObjects.UserGroups, group)
	// catching invalid kinds of objects for this apiVersion
	default:
		err = manifest.UnsupportedKindErr(o)
	}
	return err
}

// Validate performs validation of parsed APIObjects.
func (o APIObjects) Validate() (err error) {
	var errs []error
	if err = validateUniquenessConstraints(manifest.KindSLO, o.SLOs); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindService, o.Services); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindProject, o.Projects); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindAgent, o.Agents); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindDirect, o.Directs); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindAlertMethod, o.AlertMethods); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindAlertPolicy, o.AlertPolicies); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindAlertSilence, o.AlertSilences); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindDataExport, o.DataExports); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindRoleBinding, o.RoleBindings); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindAnnotation, o.Annotations); err != nil {
		errs = append(errs, err)
	}
	if err = validateUniquenessConstraints(manifest.KindUserGroup, o.UserGroups); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		builder := strings.Builder{}
		for i, err := range errs {
			builder.WriteString(err.Error())
			if i < len(errs)-1 {
				builder.WriteString("; ")
			}
		}
		return errors.New(builder.String())
	}
	return nil
}

// validateUniquenessConstraints finds conflicting objects in a Kind slice.
// It returns an error if any conflicts were encountered.
// The error informs about the cause and lists ALL conflicts.
func validateUniquenessConstraints[T manifest.Object](kind manifest.Kind, slice []T) error {
	unique := make(map[string]struct{}, len(slice))
	var details []string
	for i := range slice {
		key := slice[i].GetName()
		if v, ok := any(slice[i]).(manifest.ProjectScopedObject); ok {
			key = v.GetProject() + "_" + key
		}
		if _, conflicts := unique[key]; conflicts {
			details = append(details, conflictDetails(slice[i], kind))
			continue
		}
		unique[key] = struct{}{}
	}
	if len(details) > 0 {
		return conflictError(kind, details)
	}
	return nil
}

// conflictDetails creates a formatted string identifying a single conflict between two objects.
func conflictDetails(object manifest.Object, kind manifest.Kind) string {
	switch v := object.(type) {
	case manifest.ProjectScopedObject:
		return fmt.Sprintf(`{"Project": "%s", "%s": "%s"}`, v.GetProject(), kind, object.GetName())
	default:
		return fmt.Sprintf(`"%s"`, object.GetName())
	}
}

// conflictError formats an error returned for a specific Kind with all it's conflicts listed as a JSON array.
// nolint: stylecheck
func conflictError(kind manifest.Kind, details []string) error {
	return fmt.Errorf(`Constraint "%s" was violated due to the following conflicts: [%s]`,
		constraintDetails(kind), strings.Join(details, ", "))
}

// constraintDetails creates a formatted string specifying the constraint which was broken.
func constraintDetails(kind manifest.Kind) string {
	switch kind {
	case manifest.KindProject, manifest.KindRoleBinding, manifest.KindUserGroup:
		return fmt.Sprintf(`%s.metadata.name has to be unique`, kind)
	default:
		return fmt.Sprintf(`%s.metadata.name has to be unique across a single Project`, kind)
	}
}
