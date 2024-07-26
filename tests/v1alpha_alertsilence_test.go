//go:build e2e_test

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_AlertSilence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)

	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})
	defaultProjectService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: defaultProject,
	})

	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})
	alertPolicyExample := examplesRegistry[manifest.KindAlertPolicy][0].Example
	alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
	alertPolicy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{
		{
			Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
				Name:    alertMethod.Metadata.Name,
				Project: alertMethod.Metadata.Project,
			},
		},
	}
	defaultProjectAlertPolicy := deepCopyObject(t, alertPolicy)
	defaultProjectAlertPolicy.Metadata.Name = generateName()
	defaultProjectAlertPolicy.Metadata.Project = defaultProject

	dataSourceType := v1alpha.Datadog
	directs := filterSlice(v1alphaSLODependencyDirects(t), func(o manifest.Object) bool {
		typ, _ := o.(v1alphaDirect.Direct).Spec.GetType()
		return typ == dataSourceType
	})
	require.Len(t, directs, 1)
	direct := directs[0].(v1alphaDirect.Direct)

	slo := getExample[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			return example.(dataSourceTypeGetter).GetDataSourceType() == dataSourceType
		},
	)
	slo.Spec.AnomalyConfig = nil
	slo.Metadata.Name = generateName()
	slo.Metadata.Project = project.GetName()
	slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo.Spec.AlertPolicies = make([]string, 0)
	slo.Spec.Service = service.Metadata.Name

	defaultProjectSLO := deepCopyObject(t, slo)
	defaultProjectSLO.Metadata.Name = generateName()
	defaultProjectSLO.Metadata.Project = defaultProject
	defaultProjectSLO.Spec.AlertPolicies = []string{defaultProjectAlertPolicy.Metadata.Name}
	defaultProjectSLO.Spec.Service = defaultProjectService.Metadata.Name

	examples := examplesRegistry[manifest.KindAlertSilence]
	allObjects := make([]manifest.Object, 0, len(examples))
	allObjects = append(
		allObjects,
		project,
		service,
		defaultProjectService,
		alertMethod,
		defaultProjectAlertPolicy,
		defaultProjectSLO,
	)

	for i, example := range examples {
		silence := newV1alphaAlertSilence(t,
			v1alphaAlertSilence.Metadata{
				Name:    generateName(),
				Project: project.GetName(),
			},
			example.GetVariant(),
			example.GetSubVariant(),
		)
		// Examples have a static time set somewhere potentially in the past.
		// Since we don't return past alert silences, we need to set the start and end times
		// somewhere in the future relative to the test execution.
		futureTime := time.Now().Add(time.Hour).UTC()
		if silence.Spec.Period.StartTime != nil {
			silence.Spec.Period.StartTime = &futureTime
		}
		if silence.Spec.Period.EndTime != nil {
			endTime := futureTime.Add(time.Hour)
			silence.Spec.Period.EndTime = &endTime
		}

		if i == 0 {
			silence.Metadata.Project = defaultProject
			silence.Spec.AlertPolicy = v1alphaAlertSilence.AlertPolicySource{
				Name: defaultProjectAlertPolicy.Metadata.Name,
			}
			silence.Spec.SLO = defaultProjectSLO.Metadata.Name
		} else {
			// Generate new AlertPolicy for every silence
			// as there can only be a single silence per SLO and AlertPolicy.
			alertPolicy.Metadata.Name = generateName()
			slo.Spec.AlertPolicies = append(slo.Spec.AlertPolicies, alertPolicy.Metadata.Name)
			allObjects = append(allObjects, deepCopyObject(t, alertPolicy))

			silence.Spec.AlertPolicy = v1alphaAlertSilence.AlertPolicySource{
				Name:    alertPolicy.Metadata.Name,
				Project: alertPolicy.Metadata.Project,
			}
			silence.Spec.SLO = slo.Metadata.Name
		}
		allObjects = append(allObjects, silence)
	}
	// Add the SLO once all the AlertPolicies have been assigned to it.
	allObjects = append(allObjects, slo)

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaAlertSilence.AlertSilence](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetAlertSilencesRequest
		expected   []v1alphaAlertSilence.AlertSilence
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetAlertSilencesRequest{Project: sdk.ProjectsWildcard},
			expected:   manifest.FilterByKind[v1alphaAlertSilence.AlertSilence](allObjects),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetAlertSilencesRequest{},
			expected:   []v1alphaAlertSilence.AlertSilence{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetAlertSilencesRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetAlertSilencesRequest{
				Project: project.GetName(),
				Names:   []string{inputs[1].Metadata.Name},
			},
			expected: []v1alphaAlertSilence.AlertSilence{inputs[1]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaAlertSilences(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaAlertSilencesAreEqual)
		})
	}
}

func newV1alphaAlertSilence(
	t *testing.T,
	metadata v1alphaAlertSilence.Metadata,
	variant,
	subVariant string,
) v1alphaAlertSilence.AlertSilence {
	t.Helper()
	ap := getExample[v1alphaAlertSilence.AlertSilence](t,
		manifest.KindAlertSilence,
		func(example v1alphaExamples.Example) bool {
			return example.GetVariant() == variant && example.GetSubVariant() == subVariant
		},
	)
	ap.Spec.Description = objectDescription
	return v1alphaAlertSilence.New(metadata, ap.Spec)
}

func assertV1alphaAlertSilencesAreEqual(t *testing.T, expected, actual v1alphaAlertSilence.AlertSilence) {
	t.Helper()
	expected = deepCopyObject(t, expected)
	org, err := client.GetOrganization(context.Background())
	require.NoError(t, err)
	expected.Organization = org
	assert.NotNil(t, actual.Status)
	actual.Status = nil
	// Project is filled automatically by the API if missing.
	if expected.Spec.AlertPolicy.Project == "" {
		expected.Spec.AlertPolicy.Project = defaultProject
	}
	// Period's start and end times are filled automatically by the API if not set in some scenarios.
	isDurationOnlyDefined := expected.Spec.Period.Duration != "" && expected.Spec.Period.StartTime == nil
	isEndTimeOnlyDefined := expected.Spec.Period.EndTime != nil &&
		expected.Spec.Period.StartTime == nil &&
		expected.Spec.Period.Duration == ""
	if isDurationOnlyDefined || isEndTimeOnlyDefined {
		if assert.NotNil(t, actual.Spec.Period.StartTime) {
			assert.True(t, time.Now().After(*actual.Spec.Period.StartTime))
		}
		actual.Spec.Period.StartTime = nil
	}
	// The API looses some time precision when returning the object, thus we need to compensate.
	if expected.Spec.Period.StartTime != nil {
		*expected.Spec.Period.StartTime = expected.Spec.Period.StartTime.Truncate(time.Microsecond)
	}
	if expected.Spec.Period.EndTime != nil {
		*expected.Spec.Period.EndTime = expected.Spec.Period.EndTime.Truncate(time.Microsecond)
	}
	assert.Equal(t, expected, actual)
}
