//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

type v1MoveSLOsTestCase struct {
	setupObjects    []manifest.Object
	payload         v1.MoveSLOsRequest
	expectedObjects []manifest.Object
	expectedError   error
}

func Test_Objects_V1_MoveSLOs(t *testing.T) {
	t.Parallel()

	directProject := newV1alphaProject(t, v1alphaProject.Metadata{
		Name: e2etestutils.GenerateName(),
	})
	direct := newV1alphaDirect(t, v1alpha.Datadog, v1alphaDirect.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: directProject.GetName(),
	})
	globalDependencyObjects := []manifest.Object{directProject, direct}
	e2etestutils.V1Apply(t, globalDependencyObjects)
	assertObjectsExists(t, direct)
	t.Cleanup(func() {
		fmt.Println("Called cleanup!")
		e2etestutils.V1Delete(t, globalDependencyObjects)
	})

	tests := map[string]v1MoveSLOsTestCase{
		"move SLO to an existing Project and Service": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			newService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: newProject.GetName(),
			})

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: newProject.GetName(),
				OldProject: oldProject.GetName(),
				Service:    newService.GetName(),
			}
			movedSLO := slo
			movedSLO.Metadata.Project = newProject.GetName()
			movedSLO.Spec.Service = newService.GetName()
			return v1MoveSLOsTestCase{
				setupObjects:    []manifest.Object{oldProject, oldService, slo, newProject, newService},
				payload:         payload,
				expectedObjects: []manifest.Object{oldProject, oldService, movedSLO, newProject, newService},
			}
		}(),
		"move SLO to an existing Project and non-existing Service": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			newServiceName := e2etestutils.GenerateName()

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: newProject.GetName(),
				OldProject: oldProject.GetName(),
				Service:    newServiceName,
			}
			movedSLO := slo
			movedSLO.Metadata.Project = newProject.GetName()
			movedSLO.Spec.Service = newServiceName
			// New service should be created automatically based on the existing Service.
			newService := v1alphaService.New(
				v1alphaService.Metadata{
					Name:    newServiceName,
					Project: newProject.GetName(),
				},
				v1alphaService.Spec{
					Description: fmt.Sprintf(
						"Service created by moving '%s' SLO from '%s' Project",
						slo.GetName(), oldProject.GetName()),
				},
			)

			return v1MoveSLOsTestCase{
				setupObjects:    []manifest.Object{oldProject, oldService, slo, newProject},
				payload:         payload,
				expectedObjects: []manifest.Object{oldProject, oldService, movedSLO, newProject, newService},
			}
		}(),
		"move SLO to an existing Project and non-existing Service (no service name)": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: newProject.GetName(),
				OldProject: oldProject.GetName(),
			}
			movedSLO := slo
			movedSLO.Metadata.Project = newProject.GetName()
			// New service should be created automatically based on the existing Service.
			newService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    oldService.GetName(),
				Project: newProject.GetName(),
			})

			return v1MoveSLOsTestCase{
				setupObjects:    []manifest.Object{oldProject, oldService, slo, newProject},
				payload:         payload,
				expectedObjects: []manifest.Object{oldProject, oldService, movedSLO, newProject, newService},
			}
		}(),
		"move SLO to a non-existing Project and Service": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			newProjectName := e2etestutils.GenerateName()
			newServiceName := e2etestutils.GenerateName()

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: newProjectName,
				OldProject: oldProject.GetName(),
				Service:    newServiceName,
			}
			movedSLO := slo
			movedSLO.Metadata.Project = newProjectName
			movedSLO.Spec.Service = newServiceName
			// Both project and service should be created automatically.
			// Project should be bare-bones only with a description.
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: newProjectName})
			newProject.Metadata.Labels = nil
			newProject.Metadata.Annotations = nil
			newProject.Spec.Description = fmt.Sprintf("Project created by moving '%s' SLO from '%s' Project",
				slo.GetName(), oldProject.GetName())
			newService := v1alphaService.New(
				v1alphaService.Metadata{
					Name:    newServiceName,
					Project: newProjectName,
				},
				v1alphaService.Spec{
					Description: fmt.Sprintf(
						"Service created by moving '%s' SLO from '%s' Project",
						slo.GetName(), oldProject.GetName()),
				},
			)

			return v1MoveSLOsTestCase{
				setupObjects:    []manifest.Object{oldProject, oldService, slo},
				payload:         payload,
				expectedObjects: []manifest.Object{oldProject, oldService, movedSLO, newProject, newService},
			}
		}(),
		"validation error": {
			setupObjects: []manifest.Object{},
			payload: v1.MoveSLOsRequest{
				SLONames:   []string{},
				NewProject: "bar",
				OldProject: "baz",
			},
			expectedError: &sdk.HTTPError{
				APIErrors: sdk.APIErrors{Errors: []sdk.APIError{{
					Title: "length must be greater than or equal to 1",
					Code:  string(rules.ErrorCodeSliceMinLength),
					Source: &sdk.APIErrorSource{
						PropertyName:  "sloNames",
						PropertyValue: "[]",
					},
				}}},
				StatusCode: http.StatusBadRequest,
				Method:     http.MethodPost,
			},
		},
		"return error for an SLO with attached Alert Policies": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
			alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
			alertPolicy.Spec.AlertMethods = nil

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: e2etestutils.GenerateName(),
				OldProject: oldProject.GetName(),
				Service:    oldService.GetName(),
			}
			// Set alert policies for SLO.
			slo.Spec.AlertPolicies = []string{alertPolicy.GetName()}

			errMsg := fmt.Sprintf("cannot move %s SLO while it has assigned Alert Policies,"+
				" detach them manually or set 'detachAlertPolicies' parameter to 'true' in the request body",
				slo.GetName())
			return v1MoveSLOsTestCase{
				setupObjects: []manifest.Object{oldProject, oldService, alertPolicy, slo},
				payload:      payload,
				expectedError: &sdk.HTTPError{
					APIErrors: sdk.APIErrors{Errors: []sdk.APIError{{
						Title: fmt.Sprintf(`{"error":"%[1]s","message":"%[1]s","statusCode":400}`, errMsg),
					}}},
					StatusCode: http.StatusBadRequest,
					Method:     http.MethodPost,
				},
			}
		}(),
		"conflict when an SLO already exists in the new Project": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			sloName := e2etestutils.GenerateName()
			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			slo.Metadata.Name = sloName

			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			newService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: newProject.GetName(),
			})

			// Create an SLO with the same name in the new project
			existingSLO := newV1alphaSLOForMoveSLO(t, newProject.GetName(), newService.GetName(), direct)
			existingSLO.Metadata.Name = sloName

			payload := v1.MoveSLOsRequest{
				SLONames:   []string{slo.GetName()},
				NewProject: newProject.GetName(),
				OldProject: oldProject.GetName(),
				Service:    newService.GetName(),
			}

			errMsg := fmt.Sprintf("%s SLO already exists in %s Project", sloName, newProject.GetName())
			return v1MoveSLOsTestCase{
				setupObjects: []manifest.Object{oldProject, oldService, slo, newProject, newService, existingSLO},
				payload:      payload,
				expectedError: &sdk.HTTPError{
					APIErrors: sdk.APIErrors{Errors: []sdk.APIError{{
						Title: fmt.Sprintf(`{"error":"%[1]s","message":"%[1]s","statusCode":409}`, errMsg),
					}}},
					StatusCode: http.StatusConflict,
					Method:     http.MethodPost,
				},
			}
		}(),
		"detach alert policies from the SLO": func() v1MoveSLOsTestCase {
			oldProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			oldService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			})
			newProject := newV1alphaProject(t, v1alphaProject.Metadata{Name: e2etestutils.GenerateName()})
			newService := newV1alphaService(t, v1alphaService.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: newProject.GetName(),
			})

			alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
			alertPolicy1 := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
			alertPolicy1.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{}
			alertPolicy2 := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
				Name:    e2etestutils.GenerateName(),
				Project: oldProject.GetName(),
			}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
			alertPolicy2.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{}

			slo := newV1alphaSLOForMoveSLO(t, oldProject.GetName(), oldService.GetName(), direct)
			slo.Spec.AlertPolicies = []string{alertPolicy1.GetName(), alertPolicy2.GetName()}

			payload := v1.MoveSLOsRequest{
				SLONames:            []string{slo.GetName()},
				NewProject:          newProject.GetName(),
				OldProject:          oldProject.GetName(),
				Service:             newService.GetName(),
				DetachAlertPolicies: true,
			}
			movedSLO := slo
			movedSLO.Metadata.Project = newProject.GetName()
			movedSLO.Spec.Service = newService.GetName()
			movedSLO.Spec.AlertPolicies = nil // Alert policies should be detached.

			dependencyObjects := []manifest.Object{
				oldProject, oldService, alertPolicy1, alertPolicy2, newProject, newService,
			}
			return v1MoveSLOsTestCase{
				setupObjects:    append(dependencyObjects, slo),
				payload:         payload,
				expectedObjects: append(dependencyObjects, movedSLO),
			}
		}(),
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e2etestutils.V1Apply(t, test.setupObjects)
			t.Cleanup(func() {
				e2etestutils.V1Delete(t, uniqueObjects(t, append(test.setupObjects, test.expectedObjects...)))
			})

			err := client.Objects().V1().MoveSLOs(t.Context(), test.payload)
			if test.expectedError != nil {
				require.Error(t, err)
				var httpErr *sdk.HTTPError
				require.ErrorAs(t, err, &httpErr)
				assert.NotEmpty(t, httpErr.TraceID)
				assert.NotEmpty(t, httpErr.URL)
				// Clear values for easier comparison.
				httpErr.TraceID = ""
				httpErr.URL = ""
				assert.Equal(t, test.expectedError, httpErr)
				return
			}
			require.NoError(t, err)

			projects := getProjectNamesFromObjects(t, test.expectedObjects)
			actualObjects := getObjectsInProjects(t, t.Context(), projects)
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
	variant := e2etestutils.GetExampleObject[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			dsGetter, ok := example.(v1alphaExamples.DataSourceTypeGetter)
			return ok && dsGetter.GetDataSourceType() == directType
		},
	)
	variant.Spec.AlertPolicies = nil
	variant.Spec.AnomalyConfig = nil
	variant.Spec.Service = service
	variant.Spec.Description = e2etestutils.GetObjectDescription()
	variant.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.GetName(),
		Project: direct.GetProject(),
		Kind:    direct.GetKind(),
	}
	metadata := v1alphaSLO.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project,
	}
	return v1alphaSLO.New(metadata, variant.Spec)
}
