//go:build e2e_test

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_Direct(t *testing.T) {
	t.Parallel()
	project := generateV1alphaProject(t)
	directTypes := filterSlice(v1alpha.DataSourceTypeValues(), v1alphaDirect.IsValidDirectType)
	allObjects := make([]manifest.Object, 0, len(directTypes)+1)
	allObjects = append(allObjects, project)

	for i, typ := range directTypes {
		direct := newV1alphaDirect(t,
			typ,
			v1alphaDirect.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: fmt.Sprintf("Direct %d", i),
				Project:     project.GetName(),
			},
		)
		if i == 0 {
			direct.Metadata.Project = defaultProject
		}
		allObjects = append(allObjects, direct)
	}

	e2etestutils.V1Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaDirect.Direct](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetDirectsRequest
		expected   []v1alphaDirect.Direct
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetDirectsRequest{Project: sdk.ProjectsWildcard},
			expected:   inputs,
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetDirectsRequest{},
			expected:   []v1alphaDirect.Direct{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetDirectsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetDirectsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[3].Metadata.Name},
			},
			expected: []v1alphaDirect.Direct{inputs[3]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaDirects(t.Context(), test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaDirectsAreEqual)
		})
	}
}

func newV1alphaDirect(
	t *testing.T,
	typ v1alpha.DataSourceType,
	metadata v1alphaDirect.Metadata,
) v1alphaDirect.Direct {
	t.Helper()
	variant := e2etestutils.GetExampleObject[v1alphaDirect.Direct](t,
		manifest.KindDirect,
		e2etestutils.FilterExamplesByDataSourceType(typ),
	)
	variant.Spec.Description = e2etestutils.GetObjectDescription()
	return v1alphaDirect.New(metadata, variant.Spec)
}

func assertV1alphaDirectsAreEqual(t *testing.T, expected, actual v1alphaDirect.Direct) {
	t.Helper()
	assert.NotNil(t, actual.Status)
	typ, _ := expected.Spec.GetType()
	assert.Equal(t, typ.String(), actual.Status.DirectType)
	actual.Status = nil
	actual.Spec.Interval = nil
	actual.Spec.Timeout = nil
	actual.Spec.Jitter = nil

	expected = deepCopyObject(t, expected)
	switch typ {
	case v1alpha.AppDynamics:
		apd := expected.Spec.AppDynamics
		expected.Spec.AppDynamics.ClientID = fmt.Sprintf("%s@%s", apd.ClientName, apd.AccountName)
		expected.Spec.AppDynamics.ClientSecret = "[hidden]"
	case v1alpha.AzureMonitor:
		expected.Spec.AzureMonitor.ClientSecret = "[hidden]"
		expected.Spec.AzureMonitor.ClientID = "[hidden]"
	case v1alpha.BigQuery:
		expected.Spec.BigQuery.ServiceAccountKey = "[hidden]"
	case v1alpha.CloudWatch:
		expected.Spec.CloudWatch.RoleARN = "[hidden]"
	case v1alpha.Datadog:
		expected.Spec.Datadog.APIKey = "[hidden]"
		expected.Spec.Datadog.ApplicationKey = "[hidden]"
	case v1alpha.Dynatrace:
		expected.Spec.Dynatrace.DynatraceToken = "[hidden]"
	case v1alpha.GCM:
		expected.Spec.GCM.ServiceAccountKey = "[hidden]"
	case v1alpha.Honeycomb:
		expected.Spec.Honeycomb.APIKey = "[hidden]"
	case v1alpha.InfluxDB:
		expected.Spec.InfluxDB.APIToken = "[hidden]"
		expected.Spec.InfluxDB.OrganizationID = "[hidden]"
	case v1alpha.Instana:
		expected.Spec.Instana.APIToken = "[hidden]"
	case v1alpha.Lightstep:
		expected.Spec.Lightstep.AppToken = "[hidden]"
	case v1alpha.LogicMonitor:
		expected.Spec.LogicMonitor.AccessID = "[hidden]"
		expected.Spec.LogicMonitor.AccessKey = "[hidden]"
	case v1alpha.NewRelic:
		expected.Spec.NewRelic.InsightsQueryKey = "[hidden]"
	case v1alpha.Pingdom:
		expected.Spec.Pingdom.APIToken = "[hidden]"
	case v1alpha.Redshift:
		expected.Spec.Redshift.RoleARN = "[hidden]"
	case v1alpha.Splunk:
		expected.Spec.Splunk.AccessToken = "[hidden]"
	case v1alpha.SplunkObservability:
		expected.Spec.SplunkObservability.AccessToken = "[hidden]"
	case v1alpha.SumoLogic:
		expected.Spec.SumoLogic.AccessID = "[hidden]"
		expected.Spec.SumoLogic.AccessKey = "[hidden]"
	case v1alpha.ThousandEyes:
		expected.Spec.ThousandEyes.OauthBearerToken = "[hidden]"
	case v1alpha.AzurePrometheus:
		expected.Spec.AzurePrometheus.ClientID = "[hidden]"
		expected.Spec.AzurePrometheus.ClientSecret = "[hidden]"
	default:
		panic(fmt.Sprintf("unexpected v1alpha.DataSourceType: %#v", typ))
	}
	assert.Equal(t, expected, actual)
}
