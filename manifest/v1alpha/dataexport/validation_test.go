package dataexport

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for DataExport '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/dataexport.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	dataExport := validDataExport()
	dataExport.APIVersion = "v0.1"
	dataExport.Kind = manifest.KindProject
	dataExport.ManifestSource = "/home/me/dataexport.yaml"
	err := validate(dataExport)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, dataExport, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	dataExport := validDataExport()
	dataExport.Metadata = Metadata{
		Name:        strings.Repeat("MY DATAEXPORT", 20),
		DisplayName: strings.Repeat("my-dataexport", 20),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	dataExport.ManifestSource = "/home/me/dataexport.yaml"
	err := validate(dataExport)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, dataExport, err, 5,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeStringDNSLabel,
		},
	)
}

func TestValidate_Spec_ExportType(t *testing.T) {
	for name, spec := range map[string]Spec{
		"passes with valid GCS spec": {
			ExportType: "GCS",
			Spec: GCSDataExportSpec{
				BucketName: "my-bucket",
			},
		},
		"passes with valid Snowflake spec": {
			ExportType: "Snowflake",
			Spec: S3DataExportSpec{
				BucketName: "my-bucket",
				RoleARN:    "arn:aws:iam::123456789012:role/my-role",
			},
		},
		"passes with valid S3 spec": {
			ExportType: "S3",
			Spec: S3DataExportSpec{
				BucketName: "my-bucket",
				RoleARN:    "arn:aws:iam::123456789012:role/my-role",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			dataExport := validDataExport()
			dataExport.Spec = spec
			err := validate(dataExport)
			testutils.AssertNoError(t, dataExport, err)
		})
	}
	t.Run("fails with unsupported export type", func(t *testing.T) {
		dataExport := validDataExport()
		dataExport.Spec.ExportType = "Azure"
		err := validate(dataExport)
		testutils.AssertContainsErrors(t, dataExport, err, 1, testutils.ExpectedError{
			Prop: "spec.exportType",
			Code: rules.ErrorCodeOneOf,
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
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: rules.ErrorCodeRequired,
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
				Code: rules.ErrorCodeStringMatchRegexp,
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
				Code: rules.ErrorCodeStringLength,
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
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.spec.roleArn",
				Code: rules.ErrorCodeRequired,
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
				Code: rules.ErrorCodeStringMatchRegexp,
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
				Code: rules.ErrorCodeStringLength,
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

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		BucketName          string
	}{
		"fails with bucket name containing hyphens": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringMatchRegexp,
				},
			},
			BucketName: "My-Travel-Maps",
		},
		"fails with bucket name containing spaces": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringMatchRegexp,
				},
			},
			BucketName: "travel maps",
		},
		"fails with required bucket name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeRequired,
				},
			},
		},
		"fails with too short bucket name": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringLength,
				},
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringMatchRegexp,
				},
			},
			BucketName: "1",
		},
		"fails with too long bucket name": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringLength,
				},
				{
					Prop: "spec.spec.bucketName",
					Code: rules.ErrorCodeStringMatchRegexp,
				},
			},
			BucketName: strings.Repeat("my-bucket", 100),
		},
	} {
		t.Run(name, func(t *testing.T) {
			dataExport := validDataExport()
			dataExport.Spec.ExportType = "GCS"
			dataExport.Spec.Spec = GCSDataExportSpec{
				BucketName: test.BucketName,
			}
			err := validate(dataExport)
			testutils.AssertContainsErrors(t, dataExport, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
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
