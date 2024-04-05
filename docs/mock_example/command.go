package mock_example

import (
	"context"

	"github.com/nobl9/nobl9-go/sdk"
	"github.com/nobl9/nobl9-go/sdk/endpoints/objects"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

//go:generate go run go.uber.org/mock/mockgen -destination mocks/client.go -package mocks -typed . Client
//go:generate go run go.uber.org/mock/mockgen -destination mocks/objects_versions.go -package mocks -mock_names Versions=MockObjectsVersions -typed github.com/nobl9/nobl9-go/sdk/endpoints/objects Versions
//go:generate go run go.uber.org/mock/mockgen -destination mocks/objects_v1.go -package mocks -mock_names Endpoints=MockObjectsV1Endpoints -typed github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1 Endpoints

// Compiler check which ensures that [sdk.Client] implements the [Client] interface.
var _ Client = (*sdk.Client)(nil)

// Client is the interface for the [sdk.Client].
// It should only contain the subset of methods that are used by your code.
type Client interface {
	GetOrganization(ctx context.Context) (string, error)
	Objects() objects.Versions
}

// command operates on [Client] interface instead of concrete [sdk.Client].
type command struct {
	client Client
}

// MustGetOrganization fetches the organization and panics on error.
func (c command) MustGetOrganization(ctx context.Context) string {
	org, err := c.client.GetOrganization(ctx)
	if err != nil {
		panic(err)
	}
	return org
}

// GetProjectNames fetches all Projects and returns their names.
func (c command) GetProjectNames(ctx context.Context) ([]string, error) {
	projects, err := c.client.Objects().V1().GetV1alphaProjects(ctx, v1.GetProjectsRequest{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(projects))
	for _, project := range projects {
		names = append(names, project.GetName())
	}
	return names, nil
}
