package prometheus

import (
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/prometheus/v1"
)

//go:generate ../../../bin/ifacemaker -y " " -f ./*.go -s versions -i Versions -o versions_interface.go -p "$GOPACKAGE"

func NewVersions(apiFactory v1.APIFactory) Versions {
	return versions{apiFactory: apiFactory}
}

type versions struct {
	apiFactory v1.APIFactory
}

func (v versions) V1() v1.Endpoints {
	return v1.NewEndpoints(v.apiFactory)
}
