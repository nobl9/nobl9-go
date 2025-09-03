package v2

import (
	"github.com/nobl9/nobl9-go/manifest"
)

type ApplyRequest struct {
	Objects []manifest.Object
	DryRun  *bool
}

func (r ApplyRequest) WithDryRun(dryRun bool) ApplyRequest {
	r.DryRun = ptr(dryRun)
	return r
}

type DeleteRequest struct {
	Objects []manifest.Object
	DryRun  *bool
}

func (r DeleteRequest) WithDryRun(dryRun bool) DeleteRequest {
	r.DryRun = ptr(dryRun)
	return r
}

type DeleteByNameRequest struct {
	Kind    manifest.Kind
	Project string
	Names   []string
	DryRun  *bool
}

func (r DeleteByNameRequest) WithDryRun(dryRun bool) DeleteByNameRequest {
	r.DryRun = ptr(dryRun)
	return r
}

func ptr[T any](v T) *T { return &v }
