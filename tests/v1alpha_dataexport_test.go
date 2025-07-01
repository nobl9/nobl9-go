//go:build e2e_test

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/testutils"
)

func Test_Objects_V1_V1alpha_DataExport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	examples := testutils.GetAllExamples(t, manifest.KindDataExport)
	allObjects := make([]manifest.Object, 0, len(examples))
	allObjects = append(allObjects, project)

	for i, example := range examples {
		export := newV1alphaDataExport(t,
			v1alphaDataExport.Metadata{
				Name:        testutils.GenerateName(),
				DisplayName: fmt.Sprintf("Data Export %d", i),
				Project:     project.GetName(),
			},
			example.GetVariant(),
			example.GetSubVariant(),
		)
		if i == 0 {
			export.Metadata.Project = defaultProject
		}
		allObjects = append(allObjects, export)
	}

	testutils.V1Apply(t, allObjects)
	t.Cleanup(func() { testutils.V1Delete(t, allObjects) })
	inputs := manifest.FilterByKind[v1alphaDataExport.DataExport](allObjects)

	filterTests := map[string]struct {
		request    objectsV1.GetDataExportsRequest
		expected   []v1alphaDataExport.DataExport
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetDataExportsRequest{Project: sdk.ProjectsWildcard},
			expected:   manifest.FilterByKind[v1alphaDataExport.DataExport](allObjects),
			returnsAll: true,
		},
		"default project": {
			request:    objectsV1.GetDataExportsRequest{},
			expected:   []v1alphaDataExport.DataExport{inputs[0]},
			returnsAll: true,
		},
		"filter by project": {
			request: objectsV1.GetDataExportsRequest{
				Project: project.GetName(),
			},
			expected: inputs[1:],
		},
		"filter by name": {
			request: objectsV1.GetDataExportsRequest{
				Project: project.GetName(),
				Names:   []string{inputs[1].Metadata.Name},
			},
			expected: []v1alphaDataExport.DataExport{inputs[1]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetV1alphaDataExports(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Len(t, actual, len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaDataExportsAreEqual)
		})
	}
}

func newV1alphaDataExport(
	t *testing.T,
	metadata v1alphaDataExport.Metadata,
	variant,
	subVariant string,
) v1alphaDataExport.DataExport {
	t.Helper()
	ap := testutils.GetExampleObject[v1alphaDataExport.DataExport](t,
		manifest.KindDataExport,
		func(example v1alphaExamples.Example) bool {
			return example.GetVariant() == variant && example.GetSubVariant() == subVariant
		},
	)
	return v1alphaDataExport.New(metadata, ap.Spec)
}

func assertV1alphaDataExportsAreEqual(t *testing.T, expected, actual v1alphaDataExport.DataExport) {
	t.Helper()
	if actual.Spec.ExportType == v1alphaDataExport.DataExportTypeS3 && assert.NotNil(t, actual.Status) {
		assert.NotEmpty(t, actual.Status.AWSExternalID)
	}
	actual.Status = nil
	assert.Equal(t, expected, actual)
}
