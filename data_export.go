package nobl9

import "encoding/json"

// DataExport struct which mapped one to one with kind: DataExport yaml definition
type DataExport struct {
	ObjectHeader
	Spec   DataExportSpec   `json:"spec"`
	Status DataExportStatus `json:"status"`
}

// DataExportSpec represents content of DataExport's Spec
type DataExportSpec struct {
	ExportType string      `json:"exportType"`
	Spec       interface{} `json:"spec"`
}

// S3DataExportSpec represents content of Amazon S3 export type spec.
type S3DataExportSpec struct {
	BucketName string `json:"bucketName"`
	RoleARN    string `json:"roleArn"` //nolint:lll
}

// GCSDataExportSpec represents content of GCP Cloud Storage export type spec.
type GCSDataExportSpec struct {
	BucketName string `json:"bucketName"`
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
	ExportType string          `json:"exportType"`
	Spec       json.RawMessage `json:"spec"`
}

// genericToDataExport converts ObjectGeneric to ObjectDataExport
func genericToDataExport(o ObjectGeneric, onlyHeader bool) (DataExport, error) {
	res := DataExport{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	deg := dataExportGeneric{}
	if err := json.Unmarshal(o.Spec, &deg); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}

	resSpec := DataExportSpec{ExportType: deg.ExportType}
	switch resSpec.ExportType {
	case "S3", "Snowflake":
		resSpec.Spec = &S3DataExportSpec{}
	case "GCS":
		resSpec.Spec = &GCSDataExportSpec{}
	}
	if deg.Spec != nil {
		if err := json.Unmarshal(deg.Spec, &resSpec.Spec); err != nil {
			err = EnhanceError(o, err)
			return res, err
		}
	}
	res.Spec = resSpec
	return res, nil
}
