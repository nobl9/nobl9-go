package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	"github.com/nobl9/nobl9-go/sdk"
)

func DataExport() []Example {
	examples := []standardExample{
		{
			SubVariant: "gcs",
			Object: dataexport.New(
				dataexport.Metadata{
					Name:        "gcs-export",
					DisplayName: "Data export to Google Cloud Storage bucket",
					Project:     sdk.DefaultProject,
				},
				dataexport.Spec{
					ExportType: dataexport.DataExportTypeGCS,
					Spec: dataexport.GCSDataExportSpec{
						BucketName: "prod-data-export-bucket",
					},
				},
			),
		},
		{
			SubVariant: "s3",
			Object: dataexport.New(
				dataexport.Metadata{
					Name:        "s3-export",
					DisplayName: "Data export to AWS S3 bucket",
					Project:     sdk.DefaultProject,
				},
				dataexport.Spec{
					ExportType: dataexport.DataExportTypeS3,
					Spec: dataexport.S3DataExportSpec{
						BucketName: "data-export-bucket",
						RoleARN:    "arn:aws:iam::123456578901:role/nobl9-access",
					},
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
