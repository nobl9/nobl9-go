package labels

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

//go:generate ../../../bin/ifacemaker -y " " -f ./*.go -s versions -i Versions -o versions_interface.go -p "$GOPACKAGE"

func NewVersions(
	client endpoints.Client,
	orgGetter endpoints.OrganizationGetter,
	readObjects endpoints.ReadObjectsFunc,
	dryRun bool,
) Versions {
	return versions{
		client:      client,
		orgGetter:   orgGetter,
		readObjects: readObjects,
		dryRun:      dryRun,
	}
}

type versions struct {
	client      endpoints.Client
	orgGetter   endpoints.OrganizationGetter
	readObjects endpoints.ReadObjectsFunc
	dryRun      bool
}

func (v versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client, v.orgGetter, v.readObjects, v.dryRun)
}
