package dataexport

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_error.txt
var expectedError string

func TestValidate_AllErrors(t *testing.T) {
	err := validate(DataExport{
		Kind: manifest.KindDataExport,
		Metadata: Metadata{
			Name:        strings.Repeat("MY DATAEXPORT", 20),
			DisplayName: strings.Repeat("my-dataexport", 10),
			Project:     strings.Repeat("MY PROJECT", 20),
		},
		Spec: Spec{
			ExportType: "GCS",
			Spec: GCSDataExportSpec{
				BucketName: strings.Repeat("my-bucket", 20),
			},
		},
		ManifestSource: "/home/me/dataexport.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedError, "\n"), err.Error())
}
