package dataexport

import (
	"encoding/json"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/internal/serdeutil"
	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl DataExport

// New creates a new DataExport based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) DataExport {
	return DataExport{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindDataExport,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// DataExport struct which mapped one to one with kind: DataExport yaml definition
type DataExport struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
}

const (
	DataExportTypeS3        string = "S3"
	DataExportTypeSnowflake string = "Snowflake"
	DataExportTypeGCS       string = "GCS"
)

// Spec represents content of DataExport's Spec
type Spec struct {
	ExportType string      `json:"exportType"`
	Spec       interface{} `json:"spec"`
}

// Status represents content of Status optional for DataExport Object
type Status struct {
	ExportJob     ExportJobStatus `json:"exportJob"`
	AWSExternalID *string         `json:"awsExternalID,omitempty"`
}

// S3DataExportSpec represents content of Amazon S3 export type spec.
type S3DataExportSpec struct {
	BucketName string `json:"bucketName"`
	RoleARN    string `json:"roleArn"`
}

// GCSDataExportSpec represents content of GCP Cloud Storage export type spec.
type GCSDataExportSpec struct {
	BucketName string `json:"bucketName"`
}

// ExportJobStatus represents content of ExportJob status
type ExportJobStatus struct {
	Timestamp string `json:"timestamp,omitempty"`
	State     string `json:"state"`
}

func (s *Spec) UnmarshalYAML(data []byte) error {
	return s.unmarshalSpec(data, yaml.Unmarshal)
}

func (s *Spec) UnmarshalJSON(data []byte) error {
	return s.unmarshalSpec(data, json.Unmarshal)
}

type unmarshalFunc = func(data []byte, v any) error

func (s *Spec) unmarshalSpec(bytes []byte, unmarshal unmarshalFunc) error {
	var genericSpec struct {
		ExportType string               `json:"exportType" example:"Snowflake"`
		Spec       serdeutil.RawMessage `json:"spec"`
	}
	if err := unmarshal(bytes, &genericSpec); err != nil {
		return err
	}
	s.ExportType = genericSpec.ExportType
	switch s.ExportType {
	case DataExportTypeS3, DataExportTypeSnowflake:
		var s3Spec S3DataExportSpec
		if err := unmarshal(genericSpec.Spec, &s3Spec); err != nil {
			return err
		}
		s.Spec = s3Spec
	case DataExportTypeGCS:
		var gcsSpec GCSDataExportSpec
		if err := unmarshal(genericSpec.Spec, &gcsSpec); err != nil {
			return err
		}
		s.Spec = gcsSpec
	}
	return nil
}
