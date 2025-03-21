// Code generated by "objectimpl BudgetAdjustment"; DO NOT EDIT.

package budgetadjustment

import (
	"github.com/nobl9/govy/pkg/govy"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Ensure interfaces are implemented.
var _ manifest.Object = BudgetAdjustment{}
var _ v1alpha.ObjectContext = BudgetAdjustment{}

func (b BudgetAdjustment) GetVersion() manifest.Version {
	return b.APIVersion
}

func (b BudgetAdjustment) GetKind() manifest.Kind {
	return b.Kind
}

func (b BudgetAdjustment) GetName() string {
	return b.Metadata.Name
}

func (b BudgetAdjustment) Validate() error {
	if err := validate(b); err != nil {
		return err
	}
	return nil
}

func (b BudgetAdjustment) GetManifestSource() string {
	return b.ManifestSource
}

func (b BudgetAdjustment) SetManifestSource(src string) manifest.Object {
	b.ManifestSource = src
	return b
}

func (b BudgetAdjustment) GetOrganization() string {
	return b.Organization
}

func (b BudgetAdjustment) SetOrganization(org string) manifest.Object {
	b.Organization = org
	return b
}

func (b BudgetAdjustment) GetValidator() govy.Validator[BudgetAdjustment] {
	return validator
}
