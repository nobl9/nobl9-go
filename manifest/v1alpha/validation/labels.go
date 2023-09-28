package validation

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/labels"
)

// TODO: Maybe switch labels to use validation pkg too? Consistency...
func Labels() SingleRuleFunc[labels.Labels] {
	return func(v labels.Labels) error { return v.Validate() }
}
