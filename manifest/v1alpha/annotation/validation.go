package annotation

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var annotationValidation = validation.New[Annotation](
	validation.For(func(p Annotation) Metadata { return p.Metadata }).
		Include(metadataValidation),
	validation.For(func(p Annotation) Spec { return p.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validation.For(func(m Metadata) string { return m.Project }).
		WithName("metadata.project").
		Omitempty().
		Rules(validation.StringIsDNSSubdomain()),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Slo }).
		WithName("slo").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.ObjectiveName }).
		WithName("objectiveName").
		Omitempty().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Required().
		Rules(validation.StringLength(0, 1000)),
	validation.For(validation.GetSelf[Spec]()).
		Rules(datePropertyGreaterThanProperty(
			"endTime", func(s Spec) time.Time { return s.EndTime },
			"startTime", func(s Spec) time.Time { return s.StartTime },
		)),
)

func validate(p Annotation) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(annotationValidation, p)
}

// DatePropertyGreaterThanProperty checks if getter returned value passed as greaterGetter argument
// is greater that value returned by lowerGetter
func datePropertyGreaterThanProperty[S any](
	greaterProperty string, greaterGetter func(s S) time.Time,
	lowerProperty string, lowerGetter func(s S) time.Time,
) validation.SingleRule[S] {
	return validation.NewSingleRule(func(s S) error {
		greater := greaterGetter(s)
		lower := lowerGetter(s)

		if !greater.After(lower) {
			return errors.Errorf(
				`"%s" in property "%s" must be greater than "%s" in property "%s"`,
				greater, greaterProperty, lower, lowerProperty,
			)
		}

		return nil
	}).WithErrorCode(validation.ErrorCodeDateGreaterThan)
}
