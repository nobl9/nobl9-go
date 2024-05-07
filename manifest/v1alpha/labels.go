package v1alpha

import (
	_ "embed"
	"regexp"

	"github.com/nobl9/nobl9-go/internal/validation"
)

type (
	Labels map[labelKey][]labelValue

	labelKey   = string
	labelValue = string
)

const (
	minLabelKeyLength   = 1
	maxLabelKeyLength   = 63
	maxLabelValueLength = 200
)

//go:embed labels_examples.yaml
var labelsExamples string

var labelKeyRegexp = regexp.MustCompile(`^\p{Ll}([_\-0-9\p{Ll}]*[0-9\p{Ll}])?$`)

func LabelsValidationRules() validation.Validator[Labels] {
	return validation.New[Labels](
		validation.ForMap(validation.GetSelf[Labels]()).
			RulesForKeys(
				validation.StringLength(minLabelKeyLength, maxLabelKeyLength),
				validation.StringMatchRegexp(labelKeyRegexp),
			).
			IncludeForValues(labelValuesValidation).
			WithExamples(labelsExamples),
	)
}

var labelValuesValidation = validation.New[[]labelValue](
	validation.ForSlice(validation.GetSelf[[]labelValue]()).
		Rules(validation.SliceUnique(validation.SelfHashFunc[labelValue]())).
		RulesForEach(
			validation.StringMaxLength(maxLabelValueLength),
		),
)
