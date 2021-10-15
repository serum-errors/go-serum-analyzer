package analysis

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestSliceToSet(t *testing.T) {
	tests := []struct {
		slice []string
		set   CodeSet
	}{
		{[]string{"one"}, Set("one")},
		{[]string{"one", "two"}, Set("one", "two")},
		{[]string{"one", "one"}, Set("one")},
		{[]string{"three", "one", "two", "one"}, Set("one", "two", "three")},
	}

	for _, test := range tests {
		result := SliceToSet(test.slice)
		if !reflect.DeepEqual(test.set, result) {
			t.Errorf("sliceToSet(%v) should be %v but was %v", test.slice, test.set, result)
		}
	}
}

func TestSetToSlice(t *testing.T) {
	tests := []struct {
		slice []string
		set   CodeSet
	}{
		{[]string{"one"}, Set("one")},
		{[]string{"one", "two"}, Set("one", "two")},
		{[]string{"one", "three", "two"}, Set("one", "two", "three")},
	}

	for _, test := range tests {
		result := test.set.Slice()
		sort.Strings(result)
		if !reflect.DeepEqual(test.slice, result) {
			t.Errorf("%v.slice() should be %v but was %v", test.set, test.slice, result)
		}
	}
}

func TestSetAdd(t *testing.T) {
	s := Set("one")
	s.Add("two")

	expected := Set("one", "two")
	if !reflect.DeepEqual(expected, s) {
		t.Errorf("expected %v got %v", expected, s)
	}

	s.Add("one")
	s.Add("one")
	s.Add("two")
	s.Add("two")

	if !reflect.DeepEqual(expected, s) {
		t.Errorf("expected %v got %v", expected, s)
	}
}
func TestUnionAndDifference(t *testing.T) {
	tests := []struct {
		a, b, union, difference CodeSet
	}{
		{Set("one"), Set("two"), Set("one", "two"), Set("one")},
		{Set(), Set("one"), Set("one"), Set()},
		{Set("one"), Set("one"), Set("one"), Set()},
		{Set("one", "two"), Set("one", "two"), Set("one", "two"), Set()},
		{Set("three", "one", "two"), Set("two", "one"), Set("one", "two", "three"), Set("three")},
		{Set(), Set(), Set(), Set()},
		{Set(), Set("one"), Set("one"), Set()},
	}

	for _, test := range tests {
		params := fmt.Sprintf("%v, %v", test.a, test.b)

		if diff := Difference(test.a, test.b); !reflect.DeepEqual(test.difference, diff) {
			t.Errorf("difference(%s) should be %v but was %v", params, test.difference, diff)
		}

		if result := Union(test.a, test.b); !reflect.DeepEqual(test.union, result) {
			t.Errorf("union(%s) should be %v but was %v", params, test.union, result)
		}
	}
}
