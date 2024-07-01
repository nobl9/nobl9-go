package v1alphaExamples

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func TestExamples_Validate_SLO(t *testing.T) {
	for _, variant := range SLO() {
		v := variant.(sloVariant)
		t.Run(v.String(), func(t *testing.T) {
			assert.NoError(t, v.SLO.Validate())
		})
	}
}

func TestExamples_Validate_Project(t *testing.T) {
	for _, variant := range Project() {
		assert.NoError(t, variant.GetObject().(v1alphaProject.Project).Validate())
	}
}

func TestExamples_Validate_Service(t *testing.T) {
	for _, variant := range Service() {
		assert.NoError(t, variant.GetObject().(v1alphaService.Service).Validate())
	}
}

func TestExamples_Validate_AlertMethods(t *testing.T) {
	for _, variant := range AlertMethod() {
		assert.NoError(t, variant.GetObject().(v1alphaAlertMethod.AlertMethod).Validate())
	}
}

func TestExamples_Validate_Labels(t *testing.T) {
	for _, variant := range Labels() {
		assert.Nil(t, v1alpha.LabelsValidationRules().Validate(variant.GetObject().(v1alpha.Labels)))
	}
}

func TestExamples_Validate_MetadataAnnotations(t *testing.T) {
	for _, variant := range MetadataAnnotations() {
		assert.Nil(t,
			v1alpha.MetadataAnnotationsValidationRules().
				Validate(variant.GetObject().(v1alpha.MetadataAnnotations)))
	}
}
