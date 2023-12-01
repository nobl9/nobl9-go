package dataexport_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
)

func ExampleDataExport() {
	// Create the object:
	dataExport := dataexport.New(
		dataexport.Metadata{
			Name:        "s3-data-export",
			DisplayName: "S3 data export",
			Project:     "default",
		},
		dataexport.Spec{
			ExportType: "S3",
			Spec: dataexport.S3DataExportSpec{
				BucketName: "examplebucket",
				RoleARN:    "arn:aws:iam::341861879477:role/n9-access",
			},
		},
	)
	// Verify the object:
	if err := dataExport.Validate(); err != nil {
		log.Fatal("data export validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{dataExport}); err != nil {
		log.Fatal("failed to apply data export err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: DataExport
	// metadata:
	//   name: s3-data-export
	//   displayName: S3 data export
	//   project: default
	// spec:
	//   exportType: S3
	//   spec:
	//     bucketName: examplebucket
	//     roleArn: arn:aws:iam::341861879477:role/n9-access
	// status: null
}
