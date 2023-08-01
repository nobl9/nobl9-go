package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../scripts/generate-object-impl.go DataExport

type DataExportsSlice []DataExport

func (dataExports DataExportsSlice) Clone() DataExportsSlice {
	clone := make([]DataExport, len(dataExports))
	copy(clone, dataExports)
	return clone
}

const (
	DataExportTypeS3        string = "S3"
	DataExportTypeSnowflake string = "Snowflake"
	DataExportTypeGCS       string = "GCS"
)

// DataExport struct which mapped one to one with kind: DataExport yaml definition
type DataExport struct {
	APIVersion string             `json:"apiVersion"`
	Kind       manifest.Kind      `json:"kind"`
	Metadata   DataExportMetadata `json:"metadata"`
	Spec       DataExportSpec     `json:"spec"`
	Status     *DataExportStatus  `json:"status"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type DataExportMetadata struct {
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string `json:"project,omitempty" validate:"objectName"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// DataExportSpec represents content of DataExport's Spec
type DataExportSpec struct {
	ExportType string      `json:"exportType" validate:"required,exportType" example:"Snowflake"`
	Spec       interface{} `json:"spec" validate:"required"`
}

func (d *DataExportSpec) UnmarshalJSON(bytes []byte) error {
	var genericSpec struct {
		ExportType string          `json:"exportType" validate:"required,exportType" example:"Snowflake"`
		Spec       json.RawMessage `json:"spec"`
	}
	if err := json.Unmarshal(bytes, &genericSpec); err != nil {
		return err
	}
	d.ExportType = genericSpec.ExportType
	switch d.ExportType {
	case DataExportTypeS3, DataExportTypeSnowflake:
		d.Spec = &S3DataExportSpec{}
	case DataExportTypeGCS:
		d.Spec = &GCSDataExportSpec{}
	}
	if genericSpec.Spec != nil {
		if err := json.Unmarshal(genericSpec.Spec, &d.Spec); err != nil {
			return err
		}
	}
	return nil
}

// S3DataExportSpec represents content of Amazon S3 export type spec.
type S3DataExportSpec struct {
	BucketName string `json:"bucketName" validate:"required,min=3,max=63,s3BucketName" example:"examplebucket"`
	RoleARN    string `json:"roleArn" validate:"required,min=20,max=2048,roleARN" example:"arn:aws:iam::12345/role/n9-access"` //nolint:lll
}

// GCSDataExportSpec represents content of GCP Cloud Storage export type spec.
type GCSDataExportSpec struct {
	BucketName string `json:"bucketName" validate:"required,min=3,max=222,gcsBucketName" example:"example-bucket.org.com"`
}

// DataExportStatus represents content of Status optional for DataExport Object
type DataExportStatus struct {
	ExportJob     DataExportStatusJob `json:"exportJob"`
	AWSExternalID *string             `json:"awsExternalID,omitempty"`
}

// DataExportStatusJob represents content of ExportJob status
type DataExportStatusJob struct {
	Timestamp string `json:"timestamp,omitempty" example:"2021-02-09T10:43:07Z"`
	State     string `json:"state" example:"finished"`
}
