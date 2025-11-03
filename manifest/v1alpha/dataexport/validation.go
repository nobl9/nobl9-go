package dataexport

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	GCSNonDomainNameBucketMaxLength int = 63
)

var S3BucketNameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9\-.]{1,61}[a-z0-9]$`)
var DNSNameRegexp = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
var GCSNonDNSNameBucketNameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{1,61}[a-z0-9]$`)

func validate(s DataExport) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, s, manifest.KindDataExport)
}

var validator = govy.New[DataExport](
	validationV1Alpha.FieldRuleAPIVersion(func(d DataExport) manifest.Version { return d.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(d DataExport) manifest.Kind { return d.Kind }, manifest.KindDataExport),
	govy.For(func(d DataExport) Metadata { return d.Metadata }).
		Include(metadataValidation),
	govy.For(func(d DataExport) Spec { return d.Spec }).
		WithName("spec").
		Include(specValidation).
		Include(s3SpecValidation).
		Include(gcsSpecValidation),
)

var metadataValidation = govy.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
)

var specValidation = govy.New[Spec](
	govy.For(func(s Spec) string { return s.ExportType }).
		WithName("exportType").
		Required().
		Rules(rules.OneOf(DataExportTypeS3, DataExportTypeSnowflake, DataExportTypeGCS)),
)

var s3SpecValidation = govy.New[Spec](
	govy.For(func(s Spec) S3DataExportSpec {
		if spec, ok := s.Spec.(S3DataExportSpec); ok {
			return spec
		}
		return S3DataExportSpec{}
	}).
		WithName("spec").
		Include(s3Validation),
).
	When(
		func(s Spec) bool { return s.ExportType == DataExportTypeS3 || s.ExportType == DataExportTypeSnowflake },
		govy.WhenDescriptionf("exportType is either '%s' or '%s'",
			DataExportTypeS3, DataExportTypeSnowflake),
	)

var s3Validation = govy.New[S3DataExportSpec](
	govy.For(func(c S3DataExportSpec) string { return c.BucketName }).
		WithName("bucketName").
		Required().
		Rules(
			rules.StringMatchRegexp(S3BucketNameRegexp).
				WithDetails("must be a valid S3 bucket name")),
	govy.For(func(c S3DataExportSpec) string { return c.RoleARN }).
		WithName("roleArn").
		Required().
		Rules(
			rules.StringLength(20, 2048),
			//nolint:lll
			//cspell:ignore FFFD
			rules.StringMatchRegexp(regexp.MustCompile(`^[\x{0009}\x{000A}\x{000D}\x{0020}-\x{007E}\x{0085}\x{00A0}-\x{D7FF}\x{E000}-\x{FFFD}\x{10000}-\x{10FFFF}]+$`)).
				WithDetails("must be a valid ARN")),
)

var gcsSpecValidation = govy.New[Spec](
	govy.For(func(s Spec) GCSDataExportSpec {
		if spec, ok := s.Spec.(GCSDataExportSpec); ok {
			return spec
		}
		return GCSDataExportSpec{}
	}).
		WithName("spec").
		Include(gcsValidation),
).
	When(
		func(s Spec) bool { return s.ExportType == DataExportTypeGCS },
		govy.WhenDescriptionf("exportType is '%s'", DataExportTypeGCS),
	)

// gcsValidation checks if name matches restrictions specified
// at https://cloud.google.com/storage/docs/naming-buckets.
var gcsValidation = govy.New[GCSDataExportSpec](
	govy.For(func(c GCSDataExportSpec) string { return c.BucketName }).
		WithName("bucketName").
		Required().
		Rules(rules.StringLength(3, 222)).
		Include(bucketNonDNSNameValidation).
		Include(bucketDNSNameValidation),
)

var bucketNonDNSNameValidation = govy.New[string](
	govy.For(govy.GetSelf[string]()).
		Rules(rules.StringMatchRegexp(GCSNonDNSNameBucketNameRegexp).
			WithDetails("must be a valid GCS bucket name")),
).
	When(
		func(n string) bool { return len(n) <= GCSNonDomainNameBucketMaxLength },
		govy.WhenDescriptionf("bucketName length is less than or equal to %d",
			GCSNonDomainNameBucketMaxLength),
	)

var bucketDNSNameValidation = govy.New[string](
	govy.For(govy.GetSelf[string]()).
		Rules(rules.StringMatchRegexp(DNSNameRegexp).
			WithDetails("must be a valid GCS bucket name")),
).
	When(
		func(n string) bool { return len(n) > GCSNonDomainNameBucketMaxLength },
		govy.WhenDescriptionf("bucketName length is greater than %d", GCSNonDomainNameBucketMaxLength),
	)
