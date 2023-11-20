package validation_test

import (
	"fmt"
	"time"

	"github.com/nobl9/nobl9-go/validation"
)

type Teacher struct {
	Name     string        `json:"name"`
	Age      time.Duration `json:"age"`
	Students []Student     `json:"students"`
}

type Student struct {
	Index string `json:"index"`
}

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
