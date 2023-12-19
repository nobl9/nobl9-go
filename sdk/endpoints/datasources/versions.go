package datasources

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/sdk/endpoints/datasources/v1"
)

func NewVersions(client endpoints.Client) Versions {
	return Versions{client: client}
}

type Versions struct {
	client endpoints.Client
}

func (v Versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client)
}
