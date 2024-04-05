package mock_example

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"mock_example/mocks"

	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func TestCommand_MustGetOrganization(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	client.
		EXPECT().
		GetOrganization(gomock.Any()).
		Return("test", nil)
	cmd := command{client: client}

	org := cmd.MustGetOrganization(context.Background())
	assert.Equal(t, "test", org)
}

func TestCommand_GetProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	v1objects := mocks.NewMockObjectsV1Endpoints(ctrl)
	v1objects.
		EXPECT().
		GetV1alphaProjects(gomock.Any(), v1.GetProjectsRequest{}).
		Return([]v1alphaProject.Project{
			v1alphaProject.New(v1alphaProject.Metadata{Name: "test"}, v1alphaProject.Spec{}),
			v1alphaProject.New(v1alphaProject.Metadata{Name: "default"}, v1alphaProject.Spec{}),
		}, nil)
	versions := mocks.NewMockObjectsVersions(ctrl)
	versions.
		EXPECT().
		V1().
		Return(v1objects)
	client := mocks.NewMockClient(ctrl)
	client.
		EXPECT().
		Objects().
		Return(versions)

	cmd := command{client: client}

	names, err := cmd.GetProjectNames(context.Background())
	require.NoError(t, err)
	require.Len(t, names, 2)
	assert.Equal(t, []string{"test", "default"}, names)
}
