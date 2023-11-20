package dataexport

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data/expected_metadata_error.txt
var expectedMetadataError string

//go:embed test_data/expected_gcs_error.txt
var expectedGCSError string

//go:embed test_data/expected_s3_error.txt
var expectedS3Error string

func TestValidate_MetadataErrors(t *testing.T) {
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
				BucketName: "my-bucket",
			},
		},
		ManifestSource: "/home/me/dataexport.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedMetadataError, "\n"), err.Error())
}

func TestValidate_GCSErrors(t *testing.T) {
	err := validate(DataExport{
		Kind: manifest.KindDataExport,
		Metadata: Metadata{
			Name:    "gcs-export",
			Project: "default",
		},
		Spec: Spec{
			ExportType: "GCS",
			Spec: GCSDataExportSpec{
				BucketName: strings.Repeat("my-bucket", 20),
			},
		},
		ManifestSource: "/home/me/dataexport.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedGCSError, "\n"), err.Error())
}

func TestValidate_S3Errors(t *testing.T) {
	err := validate(DataExport{
		Kind: manifest.KindDataExport,
		Metadata: Metadata{
			Name:    "s3-export",
			Project: "default",
		},
		Spec: Spec{
			ExportType: "S3",
			Spec: S3DataExportSpec{
				BucketName: strings.Repeat("my-bucket", 20),
			},
		},
		ManifestSource: "/home/me/dataexport.yaml",
	})
	assert.Equal(t, strings.TrimSuffix(expectedS3Error, "\n"), err.Error())
}
