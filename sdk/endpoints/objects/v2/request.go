package v2

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
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

type GetAnnotationsRequest struct {
	Project    string
	Names      []string
	SLOName    string
	From       time.Time
	To         time.Time
	Categories []v1alphaAnnotation.Category
}
