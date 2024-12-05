package v2

import "github.com/nobl9/nobl9-go/manifest"

type ApplyRequest struct {
	Objects []manifest.Object
	// DryRun allows overriding the analogous [sdk.Client] setting.
	DryRun *bool
}

func (a ApplyRequest) WithDryRun(dryRun bool) ApplyRequest {
	a.DryRun = &dryRun
	return a
}

type DeleteRequest struct {
	Objects []manifest.Object
	// DryRun allows overriding the analogous [sdk.Client] setting.
	DryRun  *bool
	Cascade bool
}

func (d DeleteRequest) WithDryRun(dryRun bool) DeleteRequest {
	d.DryRun = &dryRun
	return d
}

type DeleteByNameRequest struct {
	Kind    manifest.Kind
	Project string
	Names   []string
	// DryRun allows overriding the analogous [sdk.Client] setting.
	DryRun  *bool
	Cascade bool
}

func (d DeleteByNameRequest) WithDryRun(dryRun bool) DeleteByNameRequest {
	d.DryRun = &dryRun
	return d
}
