package objects

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1alpha"
)

func NewVersions(client endpoints.Client) Versions {
	return Versions{client: client}
}

type Versions struct {
	client endpoints.Client
}

func (v Versions) V1alpha() v1alpha.Endpoints {
	return v1alpha.NewEndpoints(v.client)
}
