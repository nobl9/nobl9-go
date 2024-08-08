package v1alphaExamples

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func TestExamples_Validate_SLO(t *testing.T) {
	for _, variant := range SLO() {
		v := variant.(SloExample)
		t.Run(v.String(), func(t *testing.T) {
			assert.NoError(t, v.Slo().Validate())
		})
	}
}

func TestExamples_Validate_Project(t *testing.T) {
	for _, variant := range Project() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaProject.Project).Validate())
		})
	}
}

func TestExamples_Validate_Service(t *testing.T) {
	for _, variant := range Service() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaService.Service).Validate())
		})
	}
}

func TestExamples_Validate_AlertMethod(t *testing.T) {
	for _, variant := range AlertMethod() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaAlertMethod.AlertMethod).Validate())
		})
	}
}

func TestExamples_Validate_Agent(t *testing.T) {
	for _, variant := range Agent() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaAgent.Agent).Validate())
		})
	}
}

func TestExamples_Validate_Direct(t *testing.T) {
	for _, variant := range Direct() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaDirect.Direct).Validate())
		})
	}
}

func TestExamples_Validate_AlertPolicy(t *testing.T) {
	for _, variant := range AlertPolicy() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaAlertPolicy.AlertPolicy).Validate())
		})
	}
}

func TestExamples_Validate_AlertSilence(t *testing.T) {
	for _, variant := range AlertSilence() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaAlertSilence.AlertSilence).Validate())
		})
	}
}

func TestExamples_Validate_Annotation(t *testing.T) {
	for _, variant := range Annotation() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaAnnotation.Annotation).Validate())
		})
	}
}

func TestExamples_Validate_BudgetAdjustment(t *testing.T) {
	for _, variant := range BudgetAdjustment() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaBudgetAdjustment.BudgetAdjustment).Validate())
		})
	}
}

func TestExamples_Validate_DataExport(t *testing.T) {
	for _, variant := range DataExport() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaDataExport.DataExport).Validate())
		})
	}
}

func TestExamples_Validate_RoleBinding(t *testing.T) {
	for _, variant := range RoleBinding() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.NoError(t, variant.GetObject().(v1alphaRoleBinding.RoleBinding).Validate())
		})
	}
}

func TestExamples_Validate_Labels(t *testing.T) {
	for _, variant := range Labels() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.Nil(t, v1alpha.LabelsValidationRules().Validate(variant.GetObject().(v1alpha.Labels)))
		})
	}
}

func TestExamples_Validate_MetadataAnnotations(t *testing.T) {
	for _, variant := range MetadataAnnotations() {
		t.Run(variant.GetVariant()+" "+variant.GetSubVariant(), func(t *testing.T) {
			assert.Nil(t,
				v1alpha.MetadataAnnotationsValidationRules().
					Validate(variant.GetObject().(v1alpha.MetadataAnnotations)))
		})
	}
}
