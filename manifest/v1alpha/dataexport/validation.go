package dataexport

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

const (
	GCSNonDomainNameBucketMaxLength int = 63
)

var S3BucketNameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9\-.]{1,61}[a-z0-9]$`)
var DNSNameRegexp = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
var GCSNonDNSNameBucketNameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{1,61}[a-z0-9]$`)

var dataExportValidation = validation.New[DataExport](
	validation.For(func(d DataExport) Metadata { return d.Metadata }).
		Include(metadataValidation),
	validation.For(func(d DataExport) Spec { return d.Spec }).
		WithName("spec").
		Include(specValidation).
		Include(s3SpecValidation).
		Include(gcsSpecValidation),
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
		Include(s3Validation),
).
	When(func(s Spec) bool { return s.ExportType == DataExportTypeS3 || s.ExportType == DataExportTypeSnowflake })

var s3Validation = validation.New[S3DataExportSpec](
	validation.For(func(c S3DataExportSpec) string { return c.BucketName }).
		WithName("bucketName").
		Required().
		Rules(
			validation.StringLength(3, 63),
			validation.StringMatchRegexp(S3BucketNameRegexp).
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
		Include(bucketNonDNSNameValidation).
		Include(bucketDNSNameValidation),
)

var bucketNonDNSNameValidation = validation.New[string](
	validation.For(validation.GetSelf[string]()).
		Rules(validation.StringMatchRegexp(GCSNonDNSNameBucketNameRegexp).
			WithDetails("must be a valid GCS bucket name")),
).
	When(func(n string) bool {
		return len(n) <= GCSNonDomainNameBucketMaxLength
	})

var bucketDNSNameValidation = validation.New[string](
	validation.For(validation.GetSelf[string]()).
		Rules(validation.StringMatchRegexp(DNSNameRegexp).
			WithDetails("must be a valid GCS bucket name")),
).
	When(func(n string) bool {
		return len(n) > GCSNonDomainNameBucketMaxLength
	})

func validate(s DataExport) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(dataExportValidation, s)
}
