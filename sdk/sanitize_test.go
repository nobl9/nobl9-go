package sdk

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data/sanitize/slo-with-computed-fields.json
var sloWithComputedFields []byte

//go:embed test_data/sanitize/slo-without-computed-fields.json
var sloWithoutComputedFields string

func TestRemoveComputedFieldsFromObjects(t *testing.T) {
	objects, err := DecodeObjects(sloWithComputedFields)
	require.NoError(t, err)

	objects = RemoveComputedFieldsFromObjects(objects)

	var buf bytes.Buffer
	err = EncodeObject(objects[0], &buf, manifest.ObjectFormatJSON)
	require.NoError(t, err)

	assert.JSONEq(t, sloWithoutComputedFields, buf.String())
}
