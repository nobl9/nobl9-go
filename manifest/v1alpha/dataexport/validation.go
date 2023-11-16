package dataexport

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

const (
	S3BucketNameRegex               string = `^[a-z0-9][a-z0-9\-.]{1,61}[a-z0-9]$`
	DNSNameRegex                    string = `^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`
	GCSNonDomainNameBucketNameRegex string = `^[a-z0-9][a-z0-9-_]{1,61}[a-z0-9]$`
	GCSNonDomainNameBucketMaxLength int    = 63
)

var dataExportValidation = validation.New[DataExport](
	validation.For(func(d DataExport) Metadata { return d.Metadata }).
		Include(metadataValidation),
	validation.For(func(d DataExport) Spec { return d.Spec }).
		WithName("spec").
		Include(specValidation, s3SpecValidation, gcsSpecValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.ExportType }).
		WithName("exportType").
		Required().
		Rules(validation.OneOf(DataExportTypeS3, DataExportTypeSnowflake, DataExportTypeGCS)),
)

var s3SpecValidation = validation.New[Spec](
	validation.For(func(s Spec) S3DataExportSpec {
		if spec, ok := s.Spec.(S3DataExportSpec); ok {
			return spec
		}
		return S3DataExportSpec{}
	}).
		WithName("spec").
		Required().
		Include(s3Validation),
).
	When(func(s Spec) bool { return s.ExportType == DataExportTypeS3 || s.ExportType == DataExportTypeSnowflake })

var s3Validation = validation.New[S3DataExportSpec](
	validation.For(func(c S3DataExportSpec) string { return c.BucketName }).
		WithName("bucketName").
		Required().
		Rules(
			validation.StringLength(3, 63),
			validation.StringMatchRegexp(regexp.MustCompile(S3BucketNameRegex)).
				WithDetails("must be a valid S3 bucket name")),
	validation.For(func(c S3DataExportSpec) string { return c.RoleARN }).
		WithName("roleArn").
		Required().
		Rules(
			validation.StringLength(20, 2048),
			validation.StringMatchRegexp(regexp.MustCompile(v1alpha.RoleARNRegex)).
				WithDetails("must be a valid ARN")),
)

var gcsSpecValidation = validation.New[Spec](
	validation.For(func(s Spec) GCSDataExportSpec {
		if spec, ok := s.Spec.(GCSDataExportSpec); ok {
			return spec
		}
		return GCSDataExportSpec{}
	}).
		WithName("spec").
		Required().
		Include(gcsValidation),
).
	When(func(s Spec) bool { return s.ExportType == DataExportTypeGCS })

// gcsValidation checks if name matches restrictions specified
// at https://cloud.google.com/storage/docs/naming-buckets.
var gcsValidation = validation.New[GCSDataExportSpec](
	validation.For(func(c GCSDataExportSpec) string { return c.BucketName }).
		WithName("bucketName").
		Required().
		Rules(validation.StringLength(3, 222)).
		Include(bucketNonDomainNameValidation).
		Include(bucketDNSNameValidation),
)

var bucketNonDomainNameValidation = validation.New[string](
	validation.For(func(n string) string { return n }).
		Rules(validation.StringMatchRegexp(regexp.MustCompile(GCSNonDomainNameBucketNameRegex)).
			WithDetails("must be a valid GCS bucket name")),
).
	When(func(n string) bool {
		return len(n) <= GCSNonDomainNameBucketMaxLength
	})

var bucketDNSNameValidation = validation.New[string](
	validation.For(func(n string) string { return n }).
		Rules(validation.StringMatchRegexp(regexp.MustCompile(DNSNameRegex)).
			WithDetails("must be a valid GCS bucket name")),
).
	When(func(n string) bool {
		return len(n) > GCSNonDomainNameBucketMaxLength
	})

func validate(s DataExport) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(dataExportValidation, s)
}
