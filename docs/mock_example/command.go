package mock_example

import (
	"context"

	"github.com/nobl9/nobl9-go/sdk/endpoints/objects"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

//go:generate mockgen -destination mocks/client.go -package mocks -typed . Client
//go:generate mockgen -destination mocks/objects_versions.go -package mocks -mock_names Versions=MockObjectsVersions -typed github.com/nobl9/nobl9-go/sdk/endpoints/objects Versions
//go:generate mockgen -destination mocks/objects_v1.go -package mocks -mock_names Endpoints=MockObjectsV1Endpoints -typed github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1 Endpoints

type Client interface {
	GetOrganization(ctx context.Context) (string, error)
	Objects() objects.Versions
}

type command struct {
	client Client
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
