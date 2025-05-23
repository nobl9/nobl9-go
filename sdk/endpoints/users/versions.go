package users

import (
	"github.com/nobl9/nobl9-go/internal/endpoints"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/users/v2"
)

//go:generate ../../../bin/ifacemaker -y " " -f ./*.go -s versions -i Versions -o versions_interface.go -p "$GOPACKAGE"

func NewVersions(client endpoints.Client) Versions {
	return versions{client: client}
}

type versions struct {
	client endpoints.Client
}

func (v versions) V2() v2.Endpoints {
	return v2.NewEndpoints(v.client)
}
