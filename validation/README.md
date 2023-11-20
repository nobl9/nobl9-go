# Validation

Package validation implements a functional API for consistent, type safe validation.
It puts heavy focus on end user errors readability, providing means for construing
information-rich error messages. Validation pipeline is lazily evaluated

## Usage

Refer to [example_test.go](./example_test.go) for a runnable example.

### Legend

- [Validator](#validator)
- [Property rules](#property-rules)
    - [PropertyRules](#propertyrules) _(single property)_
    - [PropertyRulesForEach](#propertyrulesforeach) _(slice of properties)_
- [Rule](#rule)
    - [SingleRule](#singlerule)
    - [RuleSet](#ruleset)
- [Errors](#errors)
    - [ValidatorError](#validatorerror)
    - [PropertyError](#propertyerror)
    - [RuleError](#ruleerror)
- [FAQ](#faq)

### Validator

Validator aggregates [property rules](#property-rules) into a single validation scenario,
most commonly associated with an entity, like `struct`.
In order to create a new validator use `New` constructor:

[//]: # ( @formatter:off)
```go
type Teacher struct {
    Name     string
    Age      time.Duration
}

teacherValidator := validation.New(
    ... // Your property rules go here.
)
```

Validator rules can be evaluated on condition, to specify the predicate use `When` function:

```go
teacherValidator = teacherValidator.When(func(t Teacher) bool { return t.Age < 50 })
```

In the example validation for `Teacher` instance will only be evaluated
if the `Age` property is less then 50.

In order to associate `Validator` with an entity name use `WithName` function:

```go
teacherValidator = teacherValidator.WithName("teacher")
```

This will result in an error message displaying the entity name you've provided.

If any property rules fail [ValidatorError](#validatorerror) is returned.

### Property rules

When validating structured data, namely `struct`,
each structure consists of multiple properties. \
For `struct`, these will be its fields.

Most commonly, property has it's name and value. For instance:

```go
type Teacher struct {
  Name     string     `json:"name"`
  Students []Student  `json:"students"`
}

type Student struct {
  Index string
}
```

Nested properties are represented by paths where each property is delimited by `.`.
Arrays are represented by `[<index>]`. Property name should be derived from the struct
representation visible by the errors consumer, this will most likely be JSON format.
Looking at the above teacher/student example,
we can distinguish the following property paths:

- `name`
- `students`
- `students[0].Index` _(let's assume there's only a single student)_

If any rule fails [PropertyError](#propertyerror) is returned.

#### PropertyRules

`PropertyRules` aggregates [rules](#rule) for a single property. In order to define a property rule,
use either constructor:

- `For` handles the value as is.
- `ForPointer` extracts underlying pointer value. If the pointer was `nil` and `Required` 
  was not set, property validation will be skipped.
- `ForEach` creates a new `PropertyRulesForEach` instance.

#### PropertyRulesForEach

`PropertyRulesForEach` is an extension of `PropertyRules`, it provides means of defining
rules for each property in a slice through either:

- `RulesForEach`
- `IncludeForEach`

These work the same as `PropertyRules.Rules` and `PropertyRules.Include` except they are
run for each element of the slice.

```go
studentValidation := validation.New[Student]()
```

Currently it only works with slices, maps are not supported.

### Rule

Rules receive and validate a value. If a rule is not met it returns [RuleError](#ruleerror).

#### SingleRule

This is the most basic validation building block.
It's error code can be set using `WithErrorCode` function and its error message can
also be enhanced using `WithDetails` function. 
Details are delimited by `;` character.

#### RuleSet

Rule sets are used to aggregate multiple `SingleRule` into a single validation rule.
It wraps any and all errors returned from single rules in a container which is later
on unpacked. If you use either `WithErrorCode` or `WithDetails` functions, each error
will be extended with the provided details and error code.

### Errors

Each validation level defines an error which further enhances the details of what went wrong.

#### ValidatorError

Adds top level entity name, following our teacher example,
it would be simply `teacher`, although that once again depends on how your end use perceives
this entity. It wraps multiple `PropertyError`.

#### PropertyError
Adds both property name and value. Property value is converted to a string
representation. It wraps multiple `RuleError`. 

#### RuleError
The most basic building block for validation errors, associated with a single
failing `SingleRule`. It conveys an error message and `ErrorCode`.

#### Error codes

To aid the process of testing, `ErrorCode` has been introduced along with a helper functions
`WithErrorCode` to associate `Rule` with an error code and `AddCode` to associate multiple
error codes with a single `Rule`. Multiple error codes are delimited by `:`, similar to how
wrapped errors are represented in Go.

To check if `ErrorCode` is part if a given validation error, use `HasErrorCode`.

## FAQ

### Why not use existing validation library?

Existing, established solutions are mostly based on struct tags and heavily utilize reflection.
This leaves type safety an issue to be solved and handled by developers. For simple use cases,
covered by predefined validation functions, this solutions works well enough. 
However when adding custom validation rules, type casting has to be heavily utilized,
and it becomes increasingly harder to track what exactly is being validated.
Another issue is the readability of the errors, it's often hard or even impossible to shape
the error to the developer liking.
