package v1alpha

import (
	_ "embed"
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// Labels are key-value pairs that can be attached to certain objects.
// Labels are used to select and filter these objects.
type Labels map[labelKey][]labelValue
type (
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

func LabelsValidationRules() govy.Validator[Labels] {
	return govy.New[Labels](
		govy.ForMap(govy.GetSelf[Labels]()).
			RulesForKeys(
				rules.StringLength(minLabelKeyLength, maxLabelKeyLength),
				rules.StringMatchRegexp(labelKeyRegexp),
			).
			IncludeForValues(labelValuesValidation).
			WithExamples(labelsExamples),
	)
}

var labelValuesValidation = govy.New[[]labelValue](
	govy.ForSlice(govy.GetSelf[[]labelValue]()).
		Rules(rules.SliceUnique(rules.HashFuncSelf[labelValue]())).
		RulesForEach(
			rules.StringMaxLength(maxLabelValueLength),
		),
)
