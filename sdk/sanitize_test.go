package sdk

import (
	"bytes"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

//go:embed test_data/sanitize/slo-with-computed-fields.json
var sloWithComputedFields []byte

//go:embed test_data/sanitize/slo-without-computed-fields.json
var sloWithoutComputedFields string

func TestRemoveComputedFieldsFromObjects(t *testing.T) {
	objects, err := DecodeObjects(sloWithComputedFields)
	require.NoError(t, err)

	objects, err = RemoveComputedFieldsFromObjects(objects)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = EncodeObject(objects[0], &buf, manifest.ObjectFormatJSON)
	require.NoError(t, err)

	assert.JSONEq(t, sloWithoutComputedFields, buf.String())
}

func TestRemoveComputedFieldsFromObjects_errorWhenNotStruct(t *testing.T) {
	tests := []struct {
		objects     []manifest.Object
		expectedErr string
	}{
		{
			objects:     []manifest.Object{v1alpha.GenericObject{}},
			expectedErr: "unsupported object kind v1alpha.GenericObject at index 0, expected a struct",
		},
		{
			objects:     []manifest.Object{&v1alpha.GenericObject{}},
			expectedErr: "unsupported object kind *v1alpha.GenericObject at index 0, expected a struct",
		},
		{
			objects:     []manifest.Object{v1alphaSLO.SLO{}, &v1alpha.GenericObject{}},
			expectedErr: "unsupported object kind *v1alpha.GenericObject at index 1, expected a struct",
		},
		{
			objects:     []manifest.Object{&v1alphaSLO.SLO{}, v1alphaSLO.SLO{}, v1alpha.GenericObject{}},
			expectedErr: "unsupported object kind v1alpha.GenericObject at index 2, expected a struct",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.objects), func(t *testing.T) {
			_, err := RemoveComputedFieldsFromObjects(test.objects)
			require.Error(t, err)
			assert.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestRemoveComputedFieldsFromObjects_annotationReplay(t *testing.T) {
	replayStart := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
	replayEnd := time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC)
	elapsed := int64(3600)

	newAnnotation := func() v1alphaAnnotation.Annotation {
		a := v1alphaAnnotation.New(
			v1alphaAnnotation.Metadata{Name: "replay-123", Project: "default"},
			v1alphaAnnotation.Spec{
				Slo:         "my-slo",
				Description: "user note",
				StartTime:   replayStart,
				EndTime:     replayEnd,
				Category:    v1alphaAnnotation.CategoryReplay,
				CreatedBy:   "user-id",
				Replay: &v1alphaAnnotation.ReplayFacts{
					PeriodStart:        replayStart,
					PeriodEnd:          replayEnd,
					ElapsedTimeSeconds: &elapsed,
				},
			},
		)
		a.Status = &v1alphaAnnotation.Status{UpdatedAt: "2023-05-02T17:10:05Z", IsSystem: false}
		a.Organization = "my-org"
		a.ManifestSource = "/home/me/annotation.yaml"
		return a
	}

	assertSanitized := func(t *testing.T, got v1alphaAnnotation.Annotation) {
		t.Helper()
		assert.Nil(t, got.Spec.Replay, "spec.replay is computed and must be stripped")
		assert.Empty(t, got.Spec.CreatedBy, "spec.createdBy is computed and must be stripped")
		assert.Nil(t, got.Status, "status is computed and must be stripped")
		assert.Empty(t, got.Organization)
		assert.Empty(t, got.ManifestSource)
		// The user-authored spec fields survive sanitization.
		assert.Equal(t, "user note", got.Spec.Description)
		assert.Equal(t, v1alphaAnnotation.CategoryReplay, got.Spec.Category)
	}

	// RemoveComputedFieldsFromObjects mutates the objects it is given. For a pointer
	// object it strips the tagged fields in place. For a value object it wraps the value
	// in a new pointer and returns that pointer in the result slice, leaving the caller's
	// original value untouched; callers must therefore read the sanitized object from the
	// returned slice, not from the value they passed in.
	t.Run("pointer object is stripped in place", func(t *testing.T) {
		a := newAnnotation()
		objects, err := RemoveComputedFieldsFromObjects(objectsOf(&a))
		require.NoError(t, err)

		got, ok := objects[0].(*v1alphaAnnotation.Annotation)
		require.True(t, ok)
		assertSanitized(t, *got)
		assert.Nil(t, a.Spec.Replay, "the pointer target itself was mutated")
	})

	t.Run("value object is converted to a pointer in the returned slice", func(t *testing.T) {
		a := newAnnotation()
		objects, err := RemoveComputedFieldsFromObjects(objectsOf(a))
		require.NoError(t, err)

		got, ok := objects[0].(*v1alphaAnnotation.Annotation)
		require.True(t, ok, "a value object is converted to a pointer")
		assertSanitized(t, *got)
		assert.NotNil(t, a.Spec.Replay, "the caller's original value is left untouched")
	})

	t.Run("get -> sanitize -> apply drops computed fields from the encoded manifest", func(t *testing.T) {
		a := newAnnotation()
		objects, err := RemoveComputedFieldsFromObjects(objectsOf(&a))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, EncodeObject(objects[0], &buf, manifest.ObjectFormatJSON))
		out := buf.String()

		assert.NotContains(t, out, `"replay"`, "computed spec.replay must not survive to the applied manifest")
		assert.NotContains(t, out, "periodStart")
		assert.NotContains(t, out, "createdBy")
		assert.NotContains(t, out, `"status"`)
		assert.Contains(t, out, "user note", "the user-authored description survives the round-trip")
	})
}

// objectsOf collects objects into a slice. Passing the objects through this helper
// keeps the slice length opaque at the call site, which avoids a gosec G602
// false positive on the range loop inside RemoveComputedFieldsFromObjects.
func objectsOf(objects ...manifest.Object) []manifest.Object { return objects }
