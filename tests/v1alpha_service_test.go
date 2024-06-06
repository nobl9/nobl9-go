//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_Service(t *testing.T) {
	ctx := context.Background()
	inputs := []v1alphaService.Service{
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Labels:      annotateLabels(v1alpha.Labels{"team": []string{"green"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
		v1alphaService.New(
			v1alphaService.Metadata{
				Name:        generateName(),
				Labels:      annotateLabels(v1alpha.Labels{"team": []string{"orange"}}),
				Annotations: commonAnnotations,
			},
			v1alphaService.Spec{
				Description: objectDescription,
			},
		),
	}

	v1Apply(t, ctx, inputs)
	t.Cleanup(func() { v1Delete(t, ctx, inputs) })

	actual, err := client.Objects().V1().GetV1alphaServices(ctx, objectsV1.GetServicesRequest{})
	require.NoError(t, err)
	assertSubset(t, actual, inputs, assertServicesAreEqual)

	actual, err = client.Objects().V1().GetV1alphaServices(ctx, objectsV1.GetServicesRequest{
		Names: []string{inputs[0].Metadata.Name},
	})
	require.NoError(t, err)
	require.Len(t, actual, 1)
	assertServicesAreEqual(t, inputs[0], actual[0])
}

func assertServicesAreEqual(t *testing.T, expected, actual v1alphaService.Service) {
	t.Helper()
	assert.Equal(t, expected, actual)
}
