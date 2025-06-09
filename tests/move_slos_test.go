//go:build e2e_test

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/sdk/models"
)

type v1MoveSLOsTestCase struct {
	setupObjects    []manifest.Object
	payload         models.MoveSLOs
	expectedObjects []manifest.Object
	expectedError   error
}

func Test_Objects_V1_MoveSLOs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	directProject := newV1alphaProject(t, v1alphaProject.Metadata{
		Name: generateName(),
	})
	direct := newV1alphaDirect(t, v1alpha.Datadog, v1alphaDirect.Metadata{
		Name:    generateName(),
		Project: directProject.GetName(),
	})
	globalDependencyObjects := []manifest.Object{directProject, direct}
	v1Apply(t, globalDependencyObjects)
	t.Cleanup(func() { v1Delete(t, globalDependencyObjects) })

	tests := map[string]v1MoveSLOsTestCase{
		"move SLO to an existing Project and Service": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: generateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{Name: generateName(), Project: oldProject.GetName()})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: generateName()})
			newService := newV1alphaService(t, v1alphaService.Metadata{Name: generateName(), Project: newProject.GetName()})

			payload := models.MoveSLOs{
				SLONames:   []string{slo.GetName()},
				NewProject: newProject.GetName(),
				OldProject: oldProject.GetName(),
				Service:    newService.GetName(),
			}
			updatedSLO := slo
			updatedSLO.Metadata.Project = newProject.GetName()
			updatedSLO.Spec.Service = newService.GetName()
			return v1MoveSLOsTestCase{
				setupObjects:    []manifest.Object{oldProject, oldService, slo, newProject, newService},
				payload:         payload,
				expectedObjects: []manifest.Object{oldProject, oldService, updatedSLO, newProject, newService},
			}
		}(),
		//"move SLO to an existing Project and non-existing Service": {},
		//"move SLO to a non-existing Project and Service":           {},
		//"return error for an SLO with attached Alert Policies":     {},
		//"conflict when an SLO already exists in the new Project":   {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v1Apply(t, test.setupObjects)
			t.Cleanup(func() { v1Delete(t, test.setupObjects) })

			err := client.Objects().V1().MoveSLOs(ctx, test.payload)
			if test.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedError, err)
				return
			}
			require.NoError(t, err)

			projects := getProjectNamesFromObjects(t, test.expectedObjects)
			actualObjects := getObjectsInProjects(t, ctx, projects)
			require.Lenf(
				t,
				actualObjects,
				len(test.expectedObjects),
				"API returned a different number of objects in the %v Projects than expected",
				projects,
			)
			assertSubset(t, actualObjects, test.expectedObjects, objectsAreEqual)
		})
	}
}

func getProjectNamesFromObjects(t *testing.T, objects []manifest.Object) []string {
	t.Helper()
	var names []string
	for _, object := range objects {
		if object.GetKind() == manifest.KindProject {
			names = append(names, object.GetName())
		}
	}
	return names
}

func getObjectsInProjects(t *testing.T, ctx context.Context, projectNames []string) []manifest.Object {
	t.Helper()
	projects, err := client.Objects().V1().GetV1alphaProjects(ctx, v1.GetProjectsRequest{Names: projectNames})
	require.NoError(t, err)
	objects := make([]manifest.Object, 0, len(projects))
	for _, project := range projects {
		objects = append(objects, project)
		objects = append(objects, getObjectsInProject(t, ctx, project.GetName())...)
	}
	return objects
}

func getObjectsInProject(t *testing.T, ctx context.Context, project string) []manifest.Object {
	t.Helper()
	kinds := []manifest.Kind{
		manifest.KindSLO,
		manifest.KindService,
		manifest.KindAlertPolicy,
	}
	var allObjects []manifest.Object
	for _, kind := range kinds {
		objects, err := client.Objects().V1().Get(
			ctx,
			kind,
			http.Header{sdk.HeaderProject: []string{project}},
			nil,
		)
		require.NoError(t, err)
		allObjects = append(allObjects, objects...)
	}
	return allObjects
}

func newV1alphaSLOForMoveSLO(
	t *testing.T,
	project, service string,
	direct v1alphaDirect.Direct,
) v1alphaSLO.SLO {
	t.Helper()

	directType, err := direct.Spec.GetType()
	require.NoError(t, err)
	variant := getExample[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			dsGetter, ok := example.(dataSourceTypeGetter)
			return ok && dsGetter.GetDataSourceType() == directType
		},
	)
	variant.Spec.AlertPolicies = nil
	variant.Spec.AnomalyConfig = nil
	variant.Spec.Service = service
	variant.Spec.Description = objectDescription
	variant.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.GetName(),
		Project: direct.GetProject(),
		Kind:    direct.GetKind(),
	}
	metadata := v1alphaSLO.Metadata{
		Name:    generateName(),
		Project: project,
	}
	return v1alphaSLO.New(metadata, variant.Spec)
}
