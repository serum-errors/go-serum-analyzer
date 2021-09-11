package analysis

import (
	"fmt"
	"strings"
)

type codeSet map[string]struct{}

func (set codeSet) String() string {
	return fmt.Sprintf("set[%s]", strings.Join(set.slice(), ", "))
}

// sliceToSet creates a set containing all values of the given slice, removing duplicates.
// The slice is not modified.
func sliceToSet(slice []string) codeSet {
	set := make(codeSet, len(slice))
	for _, value := range slice {
		set[value] = struct{}{}
	}
	return set
}

// slice creates a slice containing all values of the given set.
// The set is not modified.
func (set codeSet) slice() []string {
	slice := make([]string, 0, len(set))
	for value := range set {
		slice = append(slice, value)
	}
	return slice
}

// add adds a value to the set.
func (set codeSet) add(value string) {
	set[value] = struct{}{}
}

// union returns a set containing all values that appear in either input set.
// The input sets cannot be used afterwards as union works inplace.
func union(set, other codeSet) codeSet {
	// Make sure we add values from the smaller into the bigger set
	if len(set) < len(other) {
		set, other = other, set
	}

	for value := range other {
		set[value] = struct{}{}
	}

	return set
}
