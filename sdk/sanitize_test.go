package sdk

import (
	"bytes"
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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
		exepctedErr string
	}{
		{
			objects:     []manifest.Object{v1alpha.GenericObject{}},
			exepctedErr: "unsupported object kind map[] at index 0, expected a struct",
		},
		{
			objects:     []manifest.Object{&v1alpha.GenericObject{}},
			exepctedErr: "unsupported object kind &map[] at index 0, expected a struct",
		},
		{
			objects:     []manifest.Object{v1alphaSLO.SLO{}, &v1alpha.GenericObject{}},
			exepctedErr: "unsupported object kind &map[] at index 1, expected a struct",
		},
		{
			objects:     []manifest.Object{&v1alphaSLO.SLO{}, v1alphaSLO.SLO{}, v1alpha.GenericObject{}},
			exepctedErr: "unsupported object kind map[] at index 2, expected a struct",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.objects), func(t *testing.T) {
			_, err := RemoveComputedFieldsFromObjects(test.objects)
			require.Error(t, err)
			assert.EqualError(t, err, test.exepctedErr)
		})
	}
}
