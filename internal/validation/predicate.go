package validation

import "fmt"

type WhenOptions struct {
	description string
}

func WhenDescription(format string, a ...interface{}) WhenOptions {
	return WhenOptions{description: fmt.Sprintf(format, a...)}
}

type Predicate[S any] func(S) bool

type predicateContainer[S any] struct {
	predicate   Predicate[S]
	description string
}

type predicateMatcher[S any] struct {
	predicates []predicateContainer[S]
}

func (p predicateMatcher[S]) when(predicate Predicate[S], opts ...WhenOptions) predicateMatcher[S] {
	container := predicateContainer[S]{predicate: predicate}
	for _, opt := range opts {
		if opt.description != "" {
			container.description = opt.description
		}
	}
	p.predicates = append(p.predicates, container)
	return p
}

func (p predicateMatcher[S]) matchPredicates(st S) bool {
	for _, predicate := range p.predicates {
		if !predicate.predicate(st) {
			return false
		}
	}
	return true
}
