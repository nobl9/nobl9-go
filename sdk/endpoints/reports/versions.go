package reports

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/reports/v1"
)

//go:generate ../../../bin/ifacemaker -y " " -f ./*.go -s versions -i Versions -o versions_interface.go -p "$GOPACKAGE"

// NewVersions returns the available Reports API versions.
func NewVersions(client endpoints.Client) Versions {
	return versions{client: client}
}

type versions struct {
	client endpoints.Client
}

func (v versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client)
}
