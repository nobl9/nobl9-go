package v2

import (
	"github.com/nobl9/nobl9-go/manifest"
)

type ApplyRequest struct {
	Objects []manifest.Object
	DryRun  bool
}

type DeleteRequest struct {
	Objects []manifest.Object
	DryRun  bool
}

type DeleteByNameRequest struct {
	Kind    manifest.Kind
	Project string
	Names   []string
	DryRun  bool
}
