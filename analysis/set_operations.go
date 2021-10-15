package analysis

import (
	"fmt"
	"strings"
)

type CodeSet map[string]struct{}

func (set CodeSet) String() string {
	return fmt.Sprintf("set[%s]", strings.Join(set.Slice(), " "))
}

// Set creates a Set using the provided values.
func Set(values ...string) CodeSet {
	return SliceToSet(values)
}

// SliceToSet creates a set containing all values of the given slice, removing duplicates.
// The slice is not modified.
func SliceToSet(slice []string) CodeSet {
	set := make(CodeSet, len(slice))
	for _, value := range slice {
		set[value] = struct{}{}
	}
	return set
}

// Slice creates a Slice containing all values of the given set.
// The set is not modified.
func (set CodeSet) Slice() []string {
	slice := make([]string, 0, len(set))
	for value := range set {
		slice = append(slice, value)
	}
	return slice
}

// Add adds a value to the set.
func (set CodeSet) Add(value string) {
	set[value] = struct{}{}
}

// Union returns a set containing all values that appear in either input set.
// The input sets are not modified.
func Union(set, other CodeSet) CodeSet {
	// Assuming no collisons, so we always either allocate the correct amount or overestimate the needs.
	result := make(CodeSet, len(set)+len(other))

	for value := range set {
		result[value] = struct{}{}
	}

	for value := range other {
		result[value] = struct{}{}
	}

	return result
}

// Difference creates a new set containing the elements of the given set (lhs),
// minus the elements in the given subtrahend (rhs).
// The input sets are not modified.
func Difference(set, subtrahend CodeSet) CodeSet {
	diff := make(CodeSet)
	for value := range set {
		if _, ok := subtrahend[value]; !ok {
			diff[value] = struct{}{}
		}
	}
	return diff
}
