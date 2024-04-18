package validation

import "fmt"

type WhenOptions struct {
	description string
}

func WhenDescription(format string, a ...interface{}) WhenOptions {
	return WhenOptions{description: fmt.Sprintf(format, a...)}
}

type Predicate[S any] func(S) bool

type predicateMatcher[S any] struct {
	predicates  []Predicate[S]
	description string
}

func (p predicateMatcher[S]) when(predicate Predicate[S], opts ...WhenOptions) predicateMatcher[S] {
	for _, opt := range opts {
		if opt.description != "" {
			p.description = opt.description
		}
	}
	p.predicates = append(p.predicates, predicate)
	return p
}

func (p predicateMatcher[S]) matchPredicates(st S) bool {
	for _, predicate := range p.predicates {
		if !predicate(st) {
			return false
		}
	}
	return true
}
