package validation_test

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nobl9/nobl9-go/validation"
)

type Teacher struct {
	Name       string        `json:"name"`
	Age        time.Duration `json:"age"`
	Students   []Student     `json:"students"`
	MiddleName *string       `json:"middleName,omitempty"`
}

type Student struct {
	Index string `json:"index"`
}

const year = 24 * 365 * time.Hour

// In order to create a new [Validator] use [New] constructor.
// Let's define simple [PropertyRules] for [Teacher.Name].
// For now, it will be always failing.
func ExampleNew() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			Rules(validation.NewSingleRule(func(name string) error { return fmt.Errorf("always fails") })),
	)

	err := v.Validate(Teacher{})
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation has failed for the following properties:
	//   - always fails
}

// To associate [Validator] with an entity name use [Validator.WithName] function.
// When any of the rules fails, the error will contain the entity name you've provided.
func ExampleValidator_WithName() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			Rules(validation.NewSingleRule(func(name string) error { return fmt.Errorf("always fails") })),
	).WithName("Teacher")

	err := v.Validate(Teacher{})
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for Teacher has failed for the following properties:
	//   - always fails
}

// You can also add [Validator] name during runtime,
// by calling [ValidatorError.WithName] function on the returned error.
//
// NOTE: We left the previous "Teacher" name assignment, to demonstrate that
// the [ValidatorError.WithName] function call will shadow it.
//
// NOTE: This would also work:
//
//	err := v.WithName("Jake").Validate(Teacher{})
//
// Validation package, aside from errors handling,
// tries to follow immutability principle. Calling any function on [Validator]
// will not change its previous declaration (unless you assign it back to 'v').
func ExampleValidatorError_WithName() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			Rules(validation.NewSingleRule(func(name string) error { return fmt.Errorf("always fails") })),
	).WithName("Teacher")

	err := v.Validate(Teacher{})
	if err != nil {
		fmt.Println(err.WithName("Jake"))
	}

	// Output:
	// Validation for Jake has failed for the following properties:
	//   - always fails
}

// [Validator] rules can be evaluated on condition, to specify the predicate use [Validator.When] function.
//
// In this example, validation for [Teacher] instance will only be evaluated
// if the [Age] property is less than 50 years.
func ExampleValidator_When() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			Rules(validation.NewSingleRule(func(name string) error { return fmt.Errorf("always fails") })),
	).
		When(func(t Teacher) bool { return t.Age < (50 * year) })

	// Prepare teachers.
	teacherTom := Teacher{
		Name: "Tom",
		Age:  51 * year,
	}
	teacherJerry := Teacher{
		Name: "Jerry",
		Age:  30 * year,
	}

	// Run validation.
	err := v.Validate(teacherTom)
	if err != nil {
		fmt.Println(err.WithName("Tom"))
	}
	err = v.Validate(teacherJerry)
	if err != nil {
		fmt.Println(err.WithName("Jerry"))
	}

	// Output:
	// Validation for Jerry has failed for the following properties:
	//   - always fails
}

// So far we've been using a very simple [PropertyRules] instance:
//
//	validation.For(func(t Teacher) string { return t.Name }).
//		Rules(validation.NewSingleRule(func(name string) error { return fmt.Errorf("always fails") }))
//
// The error message returned by this property rule does not tell us
// which property is failing.
// Let's change that by adding property name using [PropertyRules.WithName].
//
// We can also change the [Rule] to be something more real.
// Validation package comes with a number of predefined [Rule], we'll use
// [EqualTo] which accepts a single argument, value to compare with.
func ExamplePropertyRules_WithName() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Rules(validation.EqualTo("Tom")),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jake",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for Teacher has failed for the following properties:
	//   - 'name' with value 'Jake':
	//     - should be equal to 'Tom'
}

// [For] constructor creates new [PropertyRules] instance.
// It's only argument, [PropertyGetter] is used to extract the property value.
// It works fine for direct values, but falls short when working with pointers.
// Often times we use pointers to indicate that a property is optional,
// or we want to discern between nil and zero values.
// In either case we want our validation rules to work on direct values,
// not the pointer, otherwise we'd have to always check if pointer != nil.
//
// [ForPointer] constructor can be used to solve this problem and allow
// us to work with the underlying value in our rules.
// Under the hood it wraps [PropertyGetter] and safely extracts the underlying value.
// If the value was nil, it will not attempt to evaluate any rules for this property.
//
// Let's define a rule for [Teacher.MiddleName] property.
// Not everyone has to have a middle name, that's why we've defined this field
// as a pointer to string, rather than a string itself.
func ExampleForPointer() {
	v := validation.New[Teacher](
		validation.ForPointer(func(t Teacher) *string { return t.MiddleName }).
			WithName("middleName").
			Rules(validation.StringMaxLength(5)),
	).WithName("Teacher")

	middleName := "Thaddeus"
	teacher := Teacher{
		Name:       "Jake",
		Age:        51 * year,
		MiddleName: &middleName,
	}

	err := v.Validate(teacher)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for Teacher has failed for the following properties:
	//   - 'middleName' with value 'Thaddeus':
	//     - length must be less than or equal to 5
}

// If you want to access the value of the entity you're writing the [Validator] for,
// you can use [GetSelf] function which is a convenience [PropertyGetter] that returns self.
// Note that we don't call [PropertyRules.WithName] here,
// as we're comparing two properties in our top level, [Teacher] scope.
//
// You can provide your own rules using [NewSingleRule] constructor.
// It returns new [SingleRule] instance which wraps your validation function.
func ExampleGetSelf() {
	customRule := validation.NewSingleRule(func(v Teacher) error {
		return fmt.Errorf("now I have access to the whole teacher")
	})

	v := validation.New[Teacher](
		validation.For(validation.GetSelf[Teacher]()).
			Rules(customRule),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jake",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for Teacher has failed for the following properties:
	//   - now I have access to the whole teacher
}

// You can use [SingleRule.WithDetails] to add additional details to the error message.
// This allows you to extend existing rules by adding your use case context.
// Let's give a regex validation some more clarity.
func ExampleSingleRule_WithDetails() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Rules(validation.StringMatchRegexp(regexp.MustCompile("^(Tom|Jerry)$")).
				WithDetails("Teacher can be either Tom or Jerry :)")),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jake",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for Teacher has failed for the following properties:
	//   - 'name' with value 'Jake':
	//     - string does not match regular expression: '^(Tom|Jerry)$'; Teacher can be either Tom or Jerry :)
}

// When testing, it can be tedious to always rely on error messages as these can change over time.
// Enter [ErrorCode], which is a simple string type alias used to ease testing,
// but also potentially allow third parties to integrate with your validation results.
// Use [SingleRule.WithErrorCode] to associate [ErrorCode] with a [SingleRule].
// Notice that our modified version of [StringMatchRegexp] will now return a new [ErrorCode].
// Predefined rules have [ErrorCode] already associated with them.
// To view the list of predefined [ErrorCode] checkout error_codes.go file.
func ExampleSingleRule_WithErrorCode() {
	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Rules(validation.StringMatchRegexp(regexp.MustCompile("^(Tom|Jerry)$")).
				WithDetails("Teacher can be either Tom or Jerry :)").
				WithErrorCode("custom_code")),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jake",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		propertyErrors := err.Errors
		ruleErrors := propertyErrors[0].Errors
		fmt.Println(ruleErrors[0].Code)
	}

	// Output:
	// custom_code
}

// Sometimes it's useful to build a [Rule] using other rules.
// To do that we'll use [RuleSet] and [NewRuleSet] constructor.
// RuleSet is a simple container for multiple [Rule].
// It is later on unpacked and each [RuleError] is reported separately.
// When [RuleSet.WithErrorCode] or [RuleSet.WithDetails] are used,
// error code and details are added to each [RuleError].
// Note that validation package uses similar syntax to wrapped errors in Go;
// a ':' delimiter is used to chain error codes together.
func ExampleRuleSet() {
	teacherNameRule := validation.NewRuleSet[string](
		validation.StringLength(1, 5),
		validation.StringMatchRegexp(regexp.MustCompile("^(Tom|Jerry)$")).
			WithDetails("Teacher can be either Tom or Jerry :)"),
	).
		WithErrorCode("teacher_name").
		WithDetails("I will add that to both rules!")

	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Rules(teacherNameRule),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jonathan",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		propertyErrors := err.Errors
		ruleErrors := propertyErrors[0].Errors
		fmt.Printf("Error codes: %s, %s\n\n", ruleErrors[0].Code, ruleErrors[1].Code)
		fmt.Println(err)
	}

	// Output:
	// Error codes: teacher_name:string_length, teacher_name:string_match_regexp
	//
	// Validation for Teacher has failed for the following properties:
	//   - 'name' with value 'Jonathan':
	//     - length must be between 1 and 5; I will add that to both rules!
	//     - string does not match regular expression: '^(Tom|Jerry)$'; Teacher can be either Tom or Jerry :); I will add that to both rules!
}

// To inspect if an error contains a given [ErrorCode], use [HasErrorCode] function.
// This function will also return true if the expected [ErrorCode]
// is part of a chain of wrapped error codes.
// In this example we're dealing with two error code chains:
// - 'teacher_name:string_length'
// - 'teacher_name:string_match_regexp'
func ExampleHasErrorCode() {
	teacherNameRule := validation.NewRuleSet[string](
		validation.StringLength(1, 5),
		validation.StringMatchRegexp(regexp.MustCompile("^(Tom|Jerry)$")),
	).
		WithErrorCode("teacher_name")

	v := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Rules(teacherNameRule),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jonathan",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		for _, code := range []validation.ErrorCode{
			"teacher_name",
			"string_length",
			"string_match_regexp",
		} {
			if validation.HasErrorCode(err, code) {
				fmt.Println("Has error code:", code)
			}
		}
	}

	// Output:
	// Has error code: teacher_name
	// Has error code: string_length
	// Has error code: string_match_regexp
}

// Sometimes you need top level context,
// but you want to scope the error to a specific, nested property.
// One of the ways to do that is to use [NewPropertyError]
// and return [PropertyError] from your validation rule.
// Note that you can still use [ErrorCode] and pass [RuleError] to the constructor.
// You can pass any number of [RuleError].
func ExamplePropertyRules_Include() {
	v := validation.New[Teacher](
		validation.For(validation.GetSelf[Teacher]()).
			Rules(validation.NewSingleRule(func(t Teacher) error {
				if t.Name == "Jake" {
					return validation.NewPropertyError(
						"name",
						t.Name,
						validation.NewRuleError("name cannot be Jake", "error_code_jake"),
						validation.NewRuleError("you can pass me too!"))
				}
				return nil
			})),
	).WithName("Teacher")

	teacher := Teacher{
		Name: "Jake",
		Age:  51 * year,
	}

	err := v.Validate(teacher)
	if err != nil {
		propertyErrors := err.Errors
		ruleErrors := propertyErrors[0].Errors
		fmt.Printf("Error code: %s\n\n", ruleErrors[0].Code)
		fmt.Println(err)
	}

	// Output:
	// Error code: error_code_jake
	//
	// Validation for Teacher has failed for the following properties:
	//   - 'name' with value 'Jake':
	//     - name cannot be Jake
	//     - you can pass me too!
}

// Sometimes you need top level context.
//func ExampleNewPropertyError() {
//	v := validation.New[Teacher](
//		validation.For(validation.GetSelf[Teacher]()).
//			Rules(validation.NewSingleRule(func(t Teacher) error {
//				return validation.NewPropertyError(
//					"name",
//					t.Name,
//					&validation.RuleError{
//						Message: "cannot have both 'bad' and 'good' metrics defined",
//						Code:    errCodeEitherBadOrGoodCountMetric,
//					}).PrependPropertyName(validation.SliceElementName("students", i))
//			})),
//	).WithName("Teacher")
//
//	teacher := Teacher{
//		Name: "Jake",
//		Age:  51 * year,
//	}
//
//	err := v.Validate(teacher)
//	if err != nil {
//		fmt.Println(err)
//	}

// Output:
// Validation for Teacher has failed for the following properties:
//   - now I have access to the whole teacher
//}

// Bringing it all together, let's create a fully fledged [Validator] for [Teacher].
func ExampleValidator() {
	studentValidator := validation.New[Student](
		validation.For(func(s Student) string { return s.Index }).
			WithName("index").
			Rules(validation.StringLength(9, 9)),
	)
	teacherValidator := validation.New[Teacher](
		validation.For(func(t Teacher) string { return t.Name }).
			WithName("name").
			Required().
			Rules(
				validation.StringNotEmpty(),
				validation.OneOf("Jake", "George")),
		validation.ForEach(func(t Teacher) []Student { return t.Students }).
			WithName("students").
			Rules(
				validation.SliceMaxLength[[]Student](2),
				validation.SliceUnique(func(v Student) string { return v.Index })).
			IncludeForEach(studentValidator),
	).When(func(t Teacher) bool { return t.Age < 50 })

	teacher := Teacher{
		Name: "John",
		Students: []Student{
			{Index: "9182300123"},
			{Index: "918230014"},
		},
	}

	err := teacherValidator.WithName("teacher").Validate(teacher)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Validation for teacher has failed for the following properties:
	//   - 'name' with value 'John':
	//     - must be one of [Jake, George]
	//   - 'students[0].index' with value '9182300123':
	//     - length must be between 9 and 9
}
