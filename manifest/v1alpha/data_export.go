package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type DataExportsSlice []DataExport

func (dataExports DataExportsSlice) Clone() DataExportsSlice {
	clone := make([]DataExport, len(dataExports))
	copy(clone, dataExports)
	return clone
}

// DataExport struct which mapped one to one with kind: DataExport yaml definition
type DataExport struct {
	manifest.ObjectHeader
	Spec   DataExportSpec   `json:"spec"`
	Status DataExportStatus `json:"status"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (d DataExport) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Project: d.Metadata.Project, Name: d.Metadata.Name}
}

// DataExportSpec represents content of DataExport's Spec
type DataExportSpec struct {
	ExportType string      `json:"exportType" validate:"required,exportType" example:"Snowflake"`
	Spec       interface{} `json:"spec" validate:"required"`
}

const (
	DataExportTypeS3        string = "S3"
	DataExportTypeSnowflake string = "Snowflake"
	DataExportTypeGCS       string = "GCS"
)

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

// dataExportGeneric represents struct to which every DataExport is parsable.
// Specific types of DataExport have different structures as Spec.
type dataExportGeneric struct {
	ExportType string          `json:"exportType" validate:"required,exportType" example:"Snowflake"`
	Spec       json.RawMessage `json:"spec"`
}

// genericToDataExport converts ObjectGeneric to ObjectDataExport
func genericToDataExport(o manifest.ObjectGeneric, v validator, onlyHeader bool) (DataExport, error) {
	res := DataExport{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	deg := dataExportGeneric{}
	if err := json.Unmarshal(o.Spec, &deg); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}

	resSpec := DataExportSpec{ExportType: deg.ExportType}
	switch resSpec.ExportType {
	case DataExportTypeS3, DataExportTypeSnowflake:
		resSpec.Spec = &S3DataExportSpec{}
	case DataExportTypeGCS:
		resSpec.Spec = &GCSDataExportSpec{}
	}
	if deg.Spec != nil {
		if err := json.Unmarshal(deg.Spec, &resSpec.Spec); err != nil {
			err = manifest.EnhanceError(o, err)
			return res, err
		}
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	return res, nil
}
