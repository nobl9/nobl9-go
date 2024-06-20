package v1alphaExamples

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestExamples_Validate_SLO(t *testing.T) {
	variants := SLO()
	for _, variant := range variants {
		t.Run(variant.String(), func(t *testing.T) {
			assert.NoError(t, variant.SLO.Validate())
		})
	}
}

func TestExamples_Validate_Project(t *testing.T) {
	project := Project()
	assert.NoError(t, project.Validate())
}

func TestExamples_Validate_Service(t *testing.T) {
	service := Service()
	assert.NoError(t, service.Validate())
}

func TestExamples_Validate_Labels(t *testing.T) {
	labels := Labels()
	assert.Nil(t, v1alpha.LabelsValidationRules().Validate(labels))
}

func TestExamples_Validate_MetadataAnnotations(t *testing.T) {
	annotations := MetadataAnnotations()
	assert.Nil(t, v1alpha.MetadataAnnotationsValidationRules().Validate(annotations))
}
