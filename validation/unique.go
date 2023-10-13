package validation

import "fmt"

// HashFunction accepts a value and returns a comparable hash.
type HashFunction[V any, H comparable] func(v V) H

// SelfHashFunc returns a HashFunction which returns it's input value as a hash itself.
// The value must be comparable.
func SelfHashFunc[H comparable]() HashFunction[H, H] {
	return func(v H) H { return v }
}

func SliceUnique[S []V, V any, H comparable](hashFunc HashFunction[V, H]) SingleRule[S] {
	return NewSingleRule(func(slice S) error {
		unique := make(map[H]int)
		for i := range slice {
			hash := hashFunc(slice[i])
			if j, ok := unique[hash]; ok {
				return fmt.Errorf("elements are not unique, index %d collides with index %d", j, i)
			}
			unique[hash] = i
		}
		return nil
	}).WithErrorCode(ErrorCodeSliceUnique)
}
