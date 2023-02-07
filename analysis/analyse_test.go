package analysis

import (
	"fmt"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestVerifyAnalyzer(t *testing.T) {
	Analyzer.Flags.Set("strict", "true")
	dir := analysistest.TestData()
	for _, pattern := range []string{
		"001",
		"annotation",
		"docformat",
		"dotimport/inner1", "dotimport",
		"error_constructor",
		"errortypes",
		"examples",
		"field_assignment",
		"func_literal",
		"interfaces/inner1", "interfaces",
		"methods",
		"multifile",
		"multipackage/inner1", "multipackage",
		"recursion",
		"typecast",
	} {
		t.Run(pattern, func(t *testing.T) {
			pattern := pattern
			analysistest.Run(t, dir, Analyzer, pattern)
		})
	}
}

type collector struct {
	data map[string]struct{}
}

func (c *collector) Errorf(format string, args ...interface{}) {
	key := fmt.Sprintf(format, args...)
	c.data[key] = struct{}{}
}

func (c *collector) assert(t *testing.T, data ...string) {
	for _, key := range data {
		if _, ok := c.data[key]; ok {
			delete(c.data, key)
			continue
		}
		t.Errorf("expected error did not appear: %s", key)
	}
	for key := range c.data {
		t.Errorf("unexpected error: %s", key)
	}
}

//
func TestNotImplemented(t *testing.T) {
	Analyzer.Flags.Set("strict", "true")
	dir := analysistest.TestData()
	for _, testcase := range []struct {
		pattern  string
		expected []string
	}{
		{
			pattern: "dereference_assignment",
			expected: []string{
				`dereference_assignment/assign.go:7:1: unexpected diagnostic: function "DereferenceAssignment" has a mismatch of declared and actual error codes: unused codes: [other-error]`,
				`dereference_assignment/assign.go:18:1: unexpected diagnostic: function "DereferenceAssignment2" has a mismatch of declared and actual error codes: unused codes: [other-error]`,
			},
		},
		{
			pattern: "type_assertion",
			expected: []string{
				`type_assertion/assert.go:6:1: unexpected diagnostic: function "TypeAssertionMethod" has a mismatch of declared and actual error codes: unused codes: [interface-error]`,
				`type_assertion/assert.go:8:9: unexpected diagnostic: type assertion is not supported in error code analysis`,
				`type_assertion/assert.go:14:1: unexpected diagnostic: function "TypeAssertionInterface" has a mismatch of declared and actual error codes: unused codes: [empty-interface-error]`,
				`type_assertion/assert.go:15:9: unexpected diagnostic: type assertion is not supported in error code analysis`,
				`type_assertion/assert.go:21:1: unexpected diagnostic: function "TypeAssertion" has a mismatch of declared and actual error codes: unused codes: [standard-error]`,
				`type_assertion/assert.go:22:9: unexpected diagnostic: type assertion is not supported in error code analysis`,
				`type_assertion/assert.go:47:1: unexpected diagnostic: function "TypeAssertionSwitch" has a mismatch of declared and actual error codes: unused codes: [maybe-error]`,
				`type_assertion/assert.go:49:14: unexpected diagnostic: type assertion switch is not supported in error code analysis`,
				`type_assertion/assert.go:63:1: unexpected diagnostic: function "TypeAssertionSwitchFromNonError" has a mismatch of declared and actual error codes: unused codes: [empty-interface-error]`,
				`type_assertion/assert.go:65:16: unexpected diagnostic: type assertion switch is not supported in error code analysis`,
			},
		},
	} {
		t.Run(testcase.pattern, func(t *testing.T) {
			testcase := testcase
			c := &collector{data: map[string]struct{}{}}
			analysistest.Run(c, dir, Analyzer, testcase.pattern)
			c.assert(t, testcase.expected...)
		})
	}
}

func TestIsErrorCodeValid(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"error", true},
		{"valid-error", true},
		{"ValidError", true},
		{"-invalid", false},
		{"invalid-", false},
		{"3invalid", false},
		{"a", true},
		{"-", false},
		{"invalid$error", false},
		{"invalid error", false},
		{"invalid       error", false},
		{"invalid\terror", false},
		{"some-2-error", true},
	}

	for _, test := range tests {
		if isErrorCodeValid(test.code) != test.valid {
			t.Errorf("isErrorCodeValid(%q) should return %v but did not", test.code, test.valid)
		}
	}
}
