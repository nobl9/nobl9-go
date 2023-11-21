package validation_test

import (
	"fmt"
	"github.com/nobl9/nobl9-go/validation"
	"time"
)

type Teacher struct {
	Name     string        `json:"name"`
	Age      time.Duration `json:"age"`
	Students []Student     `json:"students"`
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
