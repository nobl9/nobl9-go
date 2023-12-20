package objects

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func NewVersions(
	client endpoints.Client,
	orgGetter endpoints.OrganizationGetter,
	readObjects endpoints.ReadObjectsFunc,
	dryRun bool,
) Versions {
	return Versions{
		client:      client,
		orgGetter:   orgGetter,
		readObjects: readObjects,
		dryRun:      dryRun,
	}
}

type Versions struct {
	client      endpoints.Client
	orgGetter   endpoints.OrganizationGetter
	readObjects endpoints.ReadObjectsFunc
	dryRun      bool
}

func (v Versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client, v.orgGetter, v.readObjects, v.dryRun)
}
