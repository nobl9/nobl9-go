//go:build e2e_test

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/authdata/v1"
)

var (
	timeRFC3339Regexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	userIDRegexp      = regexp.MustCompile(`[a-zA-Z0-9]{20}`)
	commonAnnotations = v1alpha.MetadataAnnotations{"origin": "sdk-e2e-test"}
)

type objectsEqualityAssertFunc[T manifest.Object] func(t *testing.T, expected, actual T)

func assertSubset[T manifest.Object](t *testing.T, actual, expected []T, f objectsEqualityAssertFunc[T]) {
	t.Helper()
	for i := range expected {
		projectScoped, isProjectScoped := any(expected[i]).(manifest.ProjectScopedObject)
		found := false
		for j := range actual {
			if actual[j].GetName() != expected[i].GetName() {
				continue
			}
			if isProjectScoped {
				v, ok := any(actual[j]).(manifest.ProjectScopedObject)
				if !ok {
					continue
				}
				if projectScoped.GetProject() != v.GetProject() {
					continue
				}
			}
			f(t, expected[i], actual[j])
			found = true
			break
		}
		if !found {
			t.Errorf("expected %T %s not found in the actual list", expected[i], expected[i].GetName())
		}
	}
}

// deepCopyObject creates a deep copy of the provided object using JSON encoding and decoding.
func deepCopyObject[T any](t *testing.T, object T) T {
	t.Helper()
	data, err := json.Marshal(object)
	require.NoError(t, err)
	var copied T
	require.NoError(t, json.Unmarshal(data, &copied))
	return copied
}

func filterSlice[T any](s []T, filter func(T) bool) []T {
	result := make([]T, 0, len(s))
	for i := range s {
		if filter(s[i]) {
			result = append(result, s[i])
		}
	}
	return result
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// tryExecuteRequest will try executing a request until a timeout occurs or the request is successful.
// See [tryExecuteRequestWhile] docs for more details.
func tryExecuteRequest[T any](t *testing.T, reqFunc func() (T, error)) (T, error) {
	t.Helper()
	return tryExecuteRequestWhile(t, reqFunc, func(error) bool { return true })
}

// tryExecuteRequestWhile will try executing a request until a timeout occurs or the request is successful.
// It will keep on trying only while the `shouldRetry` condition is met.
//
// Use it ONLY for requests which check for state that propagates asynchronously.
// For instance, an object is applied instantly,
// but it will appear in SLO Status API only after an indeterminate amount of time.
func tryExecuteRequestWhile[T any](
	t *testing.T,
	reqFunc func() (T, error),
	shouldRetry func(error) bool,
) (T, error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	var (
		response T
		err      error
	)
	for {
		select {
		case <-ticker.C:
			response, err = reqFunc()
			if err == nil || !shouldRetry(err) {
				return response, err
			}
		case <-timer.C:
			t.Error("timeout")
			return response, err
		}
	}
}

func objectsAreEqual(t *testing.T, o1, o2 manifest.Object) {
	switch v1 := o1.(type) {
	case v1alphaAgent.Agent:
		require.IsType(t, v1, o2)
		assertV1alphaAgentsAreEqual(t, v1, o2.(v1alphaAgent.Agent))
	case v1alphaAlertMethod.AlertMethod:
		require.IsType(t, v1, o2)
		assertV1alphaAlertMethodsAreEqual(t, v1, o2.(v1alphaAlertMethod.AlertMethod))
	case v1alphaAlertPolicy.AlertPolicy:
		require.IsType(t, v1, o2)
		assertV1alphaAlertPoliciesAreEqual(t, v1, o2.(v1alphaAlertPolicy.AlertPolicy))
	case v1alphaAlertSilence.AlertSilence:
		require.IsType(t, v1, o2)
		assertV1alphaAlertSilencesAreEqual(t, v1, o2.(v1alphaAlertSilence.AlertSilence))
	case v1alphaAnnotation.Annotation:
		require.IsType(t, v1, o2)
		assertV1alphaAnnotationsAreEqual(t, v1, o2.(v1alphaAnnotation.Annotation))
	case v1alphaBudgetAdjustment.BudgetAdjustment:
		require.IsType(t, v1, o2)
		assertV1alphaBudgetAdjustmentsAreEqual(t, v1, o2.(v1alphaBudgetAdjustment.BudgetAdjustment))
	case v1alphaDataExport.DataExport:
		require.IsType(t, v1, o2)
		assertV1alphaDataExportsAreEqual(t, v1, o2.(v1alphaDataExport.DataExport))
	case v1alphaDirect.Direct:
		require.IsType(t, v1, o2)
		assertV1alphaDirectsAreEqual(t, v1, o2.(v1alphaDirect.Direct))
	case v1alphaProject.Project:
		require.IsType(t, v1, o2)
		assertV1alphaProjectsAreEqual(t, v1, o2.(v1alphaProject.Project))
	case v1alphaReport.Report:
		require.IsType(t, v1, o2)
		assertV1alphaReportsAreEqual(t, v1, o2.(v1alphaReport.Report))
	case v1alphaRoleBinding.RoleBinding:
		require.IsType(t, v1, o2)
		assertV1alphaRoleBindingsAreEqual(t, v1, o2.(v1alphaRoleBinding.RoleBinding))
	case v1alphaService.Service:
		require.IsType(t, v1, o2)
		assertV1alphaServicesAreEqual(t, v1, o2.(v1alphaService.Service))
	case v1alphaSLO.SLO:
		require.IsType(t, v1, o2)
		assertV1alphaSLOsAreEqual(t, v1, o2.(v1alphaSLO.SLO))
	default:
		require.Equal(t, o1, o2, "objectsAreEqual: unhandled type %T", o1)
	}
}

func uniqueObjects[T manifest.Object](t *testing.T, objects []T) []T {
	t.Helper()

	seen := make(map[string]struct{}, len(objects))
	unique := make([]T, 0, len(objects))
	for _, obj := range objects {
		key := obj.GetKind().String() + ":" + obj.GetName()
		if v, ok := any(obj).(manifest.ProjectScopedObject); ok {
			key += ":" + v.GetProject()
		}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			unique = append(unique, obj)
		}
	}
	return unique
}

func requireObjectsExists(t *testing.T, objects ...manifest.Object) {
	t.Helper()
	if !assertObjectsExists(t, objects...) {
		t.FailNow()
	}
}

func requireObjectsNotExists(t *testing.T, objects ...manifest.Object) {
	t.Helper()
	if !assertObjectsNotExists(t, objects...) {
		t.FailNow()
	}
}

func assertObjectsExists(t *testing.T, objects ...manifest.Object) bool {
	t.Helper()
	return assertObjectsExistsOrNot(t, objects, true)
}

func assertObjectsNotExists(t *testing.T, objects ...manifest.Object) bool {
	t.Helper()
	return assertObjectsExistsOrNot(t, objects, false)
}

type objectKindAndProject struct {
	Kind    manifest.Kind
	Project string
}

func (o objectKindAndProject) String() string {
	if o.Project != "" {
		return fmt.Sprintf("Kind: '%s' in Project: '%s'", o.Kind, o.Project)
	}
	return fmt.Sprintf("Kind: '%s'", o.Kind)
}

func assertObjectsExistsOrNot(t *testing.T, objects []manifest.Object, exists bool) bool {
	t.Helper()

	ok := true
	objectNamesPerKindAndProject := groupObjectNamesByKindAndProject(objects)
	for key, names := range objectNamesPerKindAndProject {
		headers := http.Header{}
		if key.Project != "" {
			headers.Set(sdk.HeaderProject, key.Project)
		}
		params := url.Values{objectsV1.QueryKeyName: names}
		objects, err := client.Objects().V1().Get(t.Context(), key.Kind, headers, params)
		if !assert.NoError(t, err) {
			ok = false
			continue
		}
		switch exists {
		case true:
			ok = assert.Lenf(t, objects, len(names),
				"expected %d objects in response, got %d (%s)", len(names), len(objects), key) && ok
		case false:
			ok = assert.Emptyf(t, objects,
				"expected no objects in response, got %d (%s)", len(objects), key) && ok
		}
	}
	return ok
}

func groupObjectNamesByKindAndProject(objects []manifest.Object) map[objectKindAndProject][]string {
	objectNamesPerKindAndProject := map[objectKindAndProject][]string{}
	for _, object := range objects {
		key := objectKindAndProject{
			Kind: object.GetKind(),
		}
		switch object := object.(type) {
		case manifest.ProjectScopedObject:
			key.Project = object.GetProject()
		case v1alphaRoleBinding.RoleBinding:
			key.Project = object.Spec.ProjectRef
		}
		objectNamesPerKindAndProject[key] = append(objectNamesPerKindAndProject[key], object.GetName())
	}
	return objectNamesPerKindAndProject
}

func ptr[T any](v T) *T { return &v }
