package validation

type Predicate[S any] func(S) bool

type predicateMatcher[S any] struct {
	predicates []Predicate[S]
}

func (r predicateMatcher[S]) when(predicates ...Predicate[S]) predicateMatcher[S] {
	r.predicates = append(r.predicates, predicates...)
	return r
}

func (r predicateMatcher[S]) matchPredicates(st S) bool {
	for _, predicate := range r.predicates {
		if !predicate(st) {
			return false
		}
	}
	return true
}
