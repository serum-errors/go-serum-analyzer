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
		set   codeSet
	}{
		{[]string{"one"}, set("one")},
		{[]string{"one", "two"}, set("one", "two")},
		{[]string{"one", "one"}, set("one")},
		{[]string{"three", "one", "two", "one"}, set("one", "two", "three")},
	}

	for _, test := range tests {
		result := sliceToSet(test.slice)
		if !reflect.DeepEqual(test.set, result) {
			t.Errorf("sliceToSet(%v) should be %v but was %v", test.slice, test.set, result)
		}
	}
}

func TestSetToSlice(t *testing.T) {
	tests := []struct {
		slice []string
		set   codeSet
	}{
		{[]string{"one"}, set("one")},
		{[]string{"one", "two"}, set("one", "two")},
		{[]string{"one", "three", "two"}, set("one", "two", "three")},
	}

	for _, test := range tests {
		result := test.set.slice()
		sort.Strings(result)
		if !reflect.DeepEqual(test.slice, result) {
			t.Errorf("%v.slice() should be %v but was %v", test.set, test.slice, result)
		}
	}
}

func TestSetAdd(t *testing.T) {
	s := set("one")
	s.add("two")

	expected := set("one", "two")
	if !reflect.DeepEqual(expected, s) {
		t.Errorf("expected %v got %v", expected, s)
	}

	s.add("one")
	s.add("one")
	s.add("two")
	s.add("two")

	if !reflect.DeepEqual(expected, s) {
		t.Errorf("expected %v got %v", expected, s)
	}
}
func TestUnionAndDifference(t *testing.T) {
	tests := []struct {
		a, b, union, difference codeSet
	}{
		{set("one"), set("two"), set("one", "two"), set("one")},
		{set(), set("one"), set("one"), set()},
		{set("one"), set("one"), set("one"), set()},
		{set("one", "two"), set("one", "two"), set("one", "two"), set()},
		{set("three", "one", "two"), set("two", "one"), set("one", "two", "three"), set("three")},
		{set(), set(), set(), set()},
		{set(), set("one"), set("one"), set()},
	}

	for _, test := range tests {
		params := fmt.Sprintf("%v, %v", test.a, test.b)

		if diff := difference(test.a, test.b); !reflect.DeepEqual(test.difference, diff) {
			t.Errorf("difference(%s) should be %v but was %v", params, test.difference, diff)
		}

		if result := union(test.a, test.b); !reflect.DeepEqual(test.union, result) {
			t.Errorf("union(%s) should be %v but was %v", params, test.union, result)
		}
	}
}
