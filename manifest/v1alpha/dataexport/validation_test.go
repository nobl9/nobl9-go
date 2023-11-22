package dataexport

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
)

//go:embed test_data/expected_metadata_error.txt
var expectedMetadataError string

func TestValidate_Metadata(t *testing.T) {
	dataExport := validDataExport()
	dataExport.Metadata = Metadata{
		Name:        strings.Repeat("MY DATAEXPORT", 20),
		DisplayName: strings.Repeat("my-dataexport", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	dataExport.ManifestSource = "/home/me/dataexport.yaml"
	err := validate(dataExport)
	require.Error(t, err)
	assert.Equal(t, strings.TrimSuffix(expectedMetadataError, "\n"), err.Error())
}

func TestValidate_Spec_ExportType(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		err := validate(dataExport)
		testutils.AssertNoError(t, dataExport, err)
	})
	t.Run("passes", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Snowflake"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertNoError(t, dataExport, err)
	})
	t.Run("passes", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "S3"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertNoError(t, dataExport, err)
	})
	t.Run("fails with unsupported export type", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Azure"
		err := validate(dataExport)
		testutils.AssertContainsErrors(t, dataExport, err, 1, testutils.ExpectedError{
			Prop: "spec.exportType",
			Code: validation.ErrorCodeOneOf,
		})
	})
}

func TestValidate_Spec_Spec_S3(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "S3"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertNoError(t, dataExport, err)
	})
	t.Run("fails with required fields", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "S3"
		dataExport.Spec.Spec = S3DataExportSpec{}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			2,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: validation.ErrorCodeRequired,
			})
	})
	t.Run("fails with invalid bucket name", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "S3"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: strings.Repeat("my-bucket", 20),
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			1,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
	t.Run("fails with invalid role ARN", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "S3"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    strings.Repeat("role-arn", 1000),
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			1,
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: validation.ErrorCodeStringLength,
			})
	})
}

func TestValidate_Spec_Spec_Snowflake(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Snowflake"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertNoError(t, dataExport, err)
	})
	t.Run("fails with required fields", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Snowflake"
		dataExport.Spec.Spec = S3DataExportSpec{}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			2,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: validation.ErrorCodeRequired,
			})
	})
	t.Run("fails with invalid bucket name", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Snowflake"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: strings.Repeat("my-bucket", 20),
			RoleARN:    "arn:aws:iam::123456789012:role/my-role",
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			1,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
	t.Run("fails with invalid role ARN", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Snowflake"
		dataExport.Spec.Spec = S3DataExportSpec{
			BucketName: "my-bucket",
			RoleARN:    strings.Repeat("role-arn", 1000),
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			1,
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: validation.ErrorCodeStringLength,
			})
	})
}

func TestValidate_Spec_Spec_GCS(t *testing.T) {
	for name, bucketName := range map[string]string{
		"passes with valid name":   "my-travel-maps",
		"passes with guid as name": "0f75d593-8e7b-4418-a5ba-cb2970f0b91e",
	} {
		t.Run(name, func(t *testing.T) {
			dataExport := validDataExport()
			dataExport.Spec.Spec = GCSDataExportSpec{
				BucketName: bucketName,
			}
			err := validate(dataExport)
			testutils.AssertNoError(t, dataExport, err)
		})
	}
	t.Run("fails with bucket name containing hyphens", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		dataExport.Spec.Spec = GCSDataExportSpec{
			BucketName: "My-Travel-Maps",
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(t, dataExport, err, 1,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
	t.Run("fails with bucket name containing spaces", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		dataExport.Spec.Spec = GCSDataExportSpec{
			BucketName: "travel maps",
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(t, dataExport, err, 1,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
	t.Run("fails with required bucket name", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		dataExport.Spec.Spec = GCSDataExportSpec{}
		err := validate(dataExport)
		testutils.AssertContainsErrors(t, dataExport, err, 1,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeRequired,
			})
	})
	t.Run("fails with too short bucket name", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		dataExport.Spec.Spec = GCSDataExportSpec{
			BucketName: "1",
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			2,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringLength,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
	t.Run("fails with too long bucket name", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "GCS"
		dataExport.Spec.Spec = GCSDataExportSpec{
			BucketName: strings.Repeat("my-bucket", 100),
		}
		err := validate(dataExport)
		testutils.AssertContainsErrors(
			t,
			dataExport,
			err,
			2,
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringLength,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.bucketName",
				Code: validation.ErrorCodeStringMatchRegexp,
			})
	})
}

func validDataExport() DataExport {
	return New(
		Metadata{
			Name:        "my-dataexport",
			DisplayName: "my dataexport",
			Project:     "default",
		},
		Spec{
			ExportType: "GCS",
			Spec: GCSDataExportSpec{
				BucketName: "my-bucket",
			},
		},
	)
}
