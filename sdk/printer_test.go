package sdk

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

//go:embed test_data/printer/expected_objects.json
var expectedObjectsJSON string

//go:embed test_data/printer/expected_objects.yaml
var expectedObjectsYAML string

func TestPrintObjects(t *testing.T) {
	objects := []manifest.Object{
		v1alpha.GenericObject{
			"apiVersion": "v1alpha",
			"kind":       "Project",
			"metadata": map[string]interface{}{
				"name":  "test-int",
				"value": 1,
			},
		},
		v1alpha.GenericObject{
			"apiVersion": "v1alpha",
			"kind":       "Project",
			"metadata": map[string]interface{}{
				"name":  "test-float",
				"value": 2.89,
			},
		},
		v1alphaProject.New(
			v1alphaProject.Metadata{
				Name: "test-project",
			},
			v1alphaProject.Spec{},
		),
	}

	t.Run("JSON format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObjects(objects, buf, manifest.ObjectFormatJSON)
		assert.NoError(t, err)
		assert.Equal(t, expectedObjectsJSON, buf.String())
	})

	t.Run("YAML format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObjects(objects, buf, manifest.ObjectFormatYAML)
		assert.NoError(t, err)
		assert.Equal(t, expectedObjectsYAML, buf.String())
	})

	t.Run("Unsupported format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObjects(objects, buf, manifest.ObjectFormat(-1))
		assert.Error(t, err)
		assert.Equal(t, "unsupported format: ObjectFormat(-1)", err.Error())
	})
}

//go:embed test_data/printer/expected_object.json
var expectedObjectJSON string

//go:embed test_data/printer/expected_object.yaml
var expectedObjectYAML string

func TestPrintObject(t *testing.T) {
	object := v1alpha.GenericObject{
		"apiVersion": "v1alpha",
		"kind":       "Project",
		"metadata": map[string]interface{}{
			"name":  "test-int",
			"value": 1,
		},
	}

	t.Run("JSON format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObject(object, buf, manifest.ObjectFormatJSON)
		assert.NoError(t, err)
		assert.Equal(t, expectedObjectJSON, buf.String())
	})

	t.Run("YAML format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObject(object, buf, manifest.ObjectFormatYAML)
		assert.NoError(t, err)
		assert.Equal(t, expectedObjectYAML, buf.String())
	})

	t.Run("Unsupported format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := PrintObject(object, buf, manifest.ObjectFormat(-1))
		assert.Error(t, err)
		assert.Equal(t, "unsupported format: ObjectFormat(-1)", err.Error())
	})
}
