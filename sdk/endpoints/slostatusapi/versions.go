package slostatusapi

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi/v1"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi/v2"
)

//go:generate ../../../bin/ifacemaker -y " " -f ./*.go -s versions -i Versions -o versions_interface.go -p "$GOPACKAGE"

func NewVersions(client endpoints.Client) Versions {
	return versions{client: client}
}

type versions struct {
	client endpoints.Client
}

func (v versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.client)
}

func (v versions) V2() v2.Endpoints {
	return v2.NewEndpoints(v.client)
}
