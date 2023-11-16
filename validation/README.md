# Validation

Package validation implements a functional API for consistent, type safe validation.
It puts heavy focus on end user errors readability providing means for construing
information-rich error messages.

## Usage

Refer to [example_test.go](./example_test.go) for a runnable example.

Validation flow is divided into the following levels:

- [Validator](#validator)
- [Property rules](#property-rules)
    - [PropertyRules](#propertyrules) _(single property)_
    - [PropertyRulesForEach](#propertyrulesforeach) _(slice of properties)_
- [Rule](#rule)
    - [SingleRule](#singlerule)
    - [RuleSet](#ruleset)
- [Errors](#errors)

The relationship between these looks as follows:

### Validator

Validator aggregates property rules into a single validation scenario,
most commonly associated with an entity like `struct`.
In order to create a new validator run:

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

### Property rules

Property rules

#### PropertyRules

#### PropertyRulesForEach

### Rule

#### SingleRule

#### RuleSet

### Errors

## FAQ

### Why not use existing validation library?

Existing, established solutions are mostly based on struct tags and heavily utilize reflection.
This leaves type safety an issue to be solved and handled by developers. For simple use cases,
covered by predefined validation functions, this solutions works well enough. 
However when adding custom validation rules, type casting has to be heavily utilized,
and it becomes increasingly harder to track what exactly is being validated.
Another issue is the readability of the errors, it's often hard or even impossible to shape
the error to the developer liking.
