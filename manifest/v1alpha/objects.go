// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// APIVersion is a value of valid apiVersions
const APIVersion = "n9/v1alpha"

// Object defines which manifest.Object are part of the manifest.VersionV1alpha.
type Object interface {
	SLO | Project | Service | Agent | Direct | Alert | AlertMethod | AlertSilence | AlertPolicy | Annotation | RoleBinding | UserGroup
}

// FilterKind filters a slice of manifest.Object and returns a subset of objects
// with the manifest.Kind defined in the type constraint.
func FilterKind[T manifest.Object](objects []manifest.Object) []T {
	var s []T
	for i := range objects {
		v, ok := objects[i].(T)
		if ok {
			s = append(s, v)
		}
	}
	return s
}

type ObjectContext interface {
	GetOrganization() string
	SetOrganization(org string) manifest.Object
	GetManifestSource() string
	SetManifestSource(src string) manifest.Object
}

// Applying multiple Agents at once can cause timeout for whole sloctl apply command.
// This is caused by long request to Okta to create client credentials app.
// The same case is applicable for delete command.
const allowedAgentsToModify = 1

//
//// Parse takes care of all Object supported by n9/v1alpha apiVersion
//func Parse(o ObjectGeneric, parsedObjects *APIObjects, onlyHeaders bool) (err error) {
//	v := NewValidator()
//	switch o.Kind {
//	case manifest.KindSLO:
//		var slo SLO
//		slo, err = genericToSLO(o, v, onlyHeaders)
//		parsedObjects.SLOs = append(parsedObjects.SLOs, slo)
//	case manifest.KindService:
//		var service Service
//		service, err = genericToService(o, v, onlyHeaders)
//		parsedObjects.Services = append(parsedObjects.Services, service)
//	case manifest.KindAgent:
//		var agent Agent
//		if len(parsedObjects.Agents) >= allowedAgentsToModify {
//			err = EnhanceError(o, errors.New("only one Agent can be defined in this configuration"))
//		} else {
//			agent, err = genericToAgent(o, v, onlyHeaders)
//			parsedObjects.Agents = append(parsedObjects.Agents, agent)
//		}
//	case manifest.KindAlertPolicy:
//		var alertPolicy AlertPolicy
//		alertPolicy, err = genericToAlertPolicy(o, v, onlyHeaders)
//		parsedObjects.AlertPolicies = append(parsedObjects.AlertPolicies, alertPolicy)
//	case manifest.KindAlertSilence:
//		var alertSilence AlertSilence
//		alertSilence, err = genericToAlertSilence(o, v, onlyHeaders)
//		parsedObjects.AlertSilences = append(parsedObjects.AlertSilences, alertSilence)
//	case manifest.KindAlertMethod:
//		var alertMethod AlertMethod
//		alertMethod, err = genericToAlertMethod(o, v, onlyHeaders)
//		parsedObjects.AlertMethods = append(parsedObjects.AlertMethods, alertMethod)
//	case manifest.KindDirect:
//		var direct Direct
//		direct, err = genericToDirect(o, v, onlyHeaders)
//		parsedObjects.Directs = append(parsedObjects.Directs, direct)
//	case manifest.KindDataExport:
//		var dataExport DataExport
//		dataExport, err = genericToDataExport(o, v, onlyHeaders)
//		parsedObjects.DataExports = append(parsedObjects.DataExports, dataExport)
//	case manifest.KindProject:
//		var project Project
//		project, err = genericToProject(o, v, onlyHeaders)
//		parsedObjects.Projects = append(parsedObjects.Projects, project)
//	case manifest.KindRoleBinding:
//		var roleBinding RoleBinding
//		roleBinding, err = genericToRoleBinding(o, v)
//		parsedObjects.RoleBindings = append(parsedObjects.RoleBindings, roleBinding)
//	case manifest.KindAnnotation:
//		var annotation Annotation
//		annotation, err = genericToAnnotation(o, v)
//		parsedObjects.Annotations = append(parsedObjects.Annotations, annotation)
//	case manifest.KindUserGroup:
//		var group UserGroup
//		group, err = genericToUserGroup(o)
//		parsedObjects.UserGroups = append(parsedObjects.UserGroups, group)
//	// catching invalid kinds of objects for this apiVersion
//	default:
//		err = UnsupportedKindErr(o)
//	}
//	return err
//}

// CheckObjectsUniqueness performs validation of parsed APIObjects.
func CheckObjectsUniqueness(objects []manifest.Object) (err error) {
	type uniqueKey struct {
		Kind    manifest.Kind
		Name    string
		Project string
	}

	unique := make(map[uniqueKey]struct{}, len(objects))
	details := make(map[manifest.Kind][]string)
	for _, obj := range objects {
		key := uniqueKey{
			Kind: obj.GetKind(),
			Name: obj.GetName(),
		}
		if v, ok := obj.(manifest.ProjectScopedObject); ok {
			key.Project = v.GetProject()
		}
		if _, conflicts := unique[key]; conflicts {
			details[obj.GetKind()] = append(details[obj.GetKind()], conflictDetails(obj, obj.GetKind()))
			continue
		}
		unique[key] = struct{}{}
	}
	var errs []error
	if len(details) > 0 {
		for kind, d := range details {
			errs = append(errs, conflictError(kind, d))
		}
	}
	if len(errs) > 0 {
		sort.Slice(errs, func(i, j int) bool { return errs[j].Error() > errs[i].Error() })
		builder := strings.Builder{}
		for i, e := range errs {
			builder.WriteString(e.Error())
			if i < len(errs)-1 {
				builder.WriteString("; ")
			}
		}
		return errors.New(builder.String())
	}
	return nil
}

// conflictDetails creates a formatted string identifying a single conflict between two objects.
func conflictDetails(object manifest.Object, kind manifest.Kind) string {
	switch v := any(object).(type) {
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
