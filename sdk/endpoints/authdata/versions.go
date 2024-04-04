package authdata

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/authdata/v1"
)

type Versions interface {
	V1() v1.Endpoints
}

func NewVersions(client endpoints.Client) Versions {
	return versions{client: client}
}

type versions struct {
	client endpoints.Client
}

func (v versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client)
}
